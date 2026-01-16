<script lang="ts">
	import SignInWrapper from '$lib/components/login-wrapper.svelte';
	import { Button } from '$lib/components/ui/button';
	import { m } from '$lib/paraglide/messages';
	import WebAuthnService from '$lib/services/webauthn-service';
	import userStore from '$lib/stores/user-store';
	import { getWebauthnErrorMessage } from '$lib/utils/error-util';
	import { startAuthentication } from '@simplewebauthn/browser';
	import { onMount } from 'svelte';
	import LoginLogoErrorSuccessIndicator from '../components/login-logo-error-success-indicator.svelte';

	let { data } = $props();

	const webauthnService = new WebAuthnService();

	let requesterIp = $state('');
	let requesterAgent = $state('');
	let expiresAt: string | null = $state(null);
	let isLoading = $state(true);
	let success = $state(false);
	let error: string | undefined = $state();

	onMount(async () => {
		if (!data.code) {
			error = 'Cross-device login link is invalid.';
			isLoading = false;
			return;
		}

		try {
			const res = await webauthnService.getCrossDeviceLoginOptions(data.code);
			requesterIp = res.requesterIp;
			requesterAgent = res.requesterUserAgent ?? res.requesterAgent ?? '';
			expiresAt = res.expiresAt;

			const authResponse = await startAuthentication({ optionsJSON: res.response });
			const user = await webauthnService.finishCrossDeviceLogin(data.code, authResponse);
			await userStore.setUser(user);
			success = true;
		} catch (e) {
			error = getWebauthnErrorMessage(e);
		} finally {
			isLoading = false;
		}
	});

	const timeLeftSeconds = $derived(() => {
		if (!expiresAt) return 0;
		const expires = new Date(expiresAt).getTime();
		return Math.max(0, Math.floor((expires - Date.now()) / 1000));
	});
</script>

<svelte:head>
	<title>{m.sign_in()}</title>
</svelte:head>

<SignInWrapper>
	<div class="flex justify-center">
		<LoginLogoErrorSuccessIndicator error={!!error} {success} />
	</div>

	{#if success}
		<h1 class="font-playfair mt-5 text-4xl font-bold">{m.sign_in()}</h1>
		<p class="text-muted-foreground mt-2">Authentication complete. You can return to your requester device.</p>
		<div class="mt-6 flex flex-col items-center gap-2 text-sm text-muted-foreground">
			<p>Requester IP: {requesterIp}</p>
			<p>Requester browser: {requesterAgent}</p>
		</div>
	{:else if error}
		<h1 class="font-playfair mt-5 text-4xl font-bold">{m.sign_in()}</h1>
		<p class="text-muted-foreground mt-2">{error}</p>
		<Button class="mt-6" href={'/login'}>
			{m.go_back()}
		</Button>
	{:else}
		<h1 class="font-playfair mt-5 text-4xl font-bold">{m.sign_in()}</h1>
		<p class="text-muted-foreground mt-2">Approve the sign in on this device.</p>
		<div class="mt-6 flex flex-col items-center gap-2 text-sm text-muted-foreground">
			<p class="text-center">
				Cross-device login initiated by<br />
				<span class="font-semibold">Requester IP: {requesterIp}</span><br />
				<span class="font-semibold">Browser: {requesterAgent}</span>
			</p>
			{#if expiresAt}
				<p class="mt-3">Expires in {timeLeftSeconds}s</p>
			{/if}
		</div>
		{#if isLoading}
			<p class="mt-6 text-sm text-muted-foreground">Waiting for your passkey…</p>
		{/if}
	{/if}
</SignInWrapper>
