<script lang="ts">
	import { afterNavigate, goto } from '$app/navigation';
	import SignInWrapper from '$lib/components/login-wrapper.svelte';
	import { Button } from '$lib/components/ui/button';
	import { m } from '$lib/paraglide/messages';
	import WebAuthnService from '$lib/services/webauthn-service';
	import userStore from '$lib/stores/user-store';
	import { getWebauthnErrorMessage } from '$lib/utils/error-util';
	import { preventDefault } from '$lib/utils/event-util';
	import { onDestroy, onMount } from 'svelte';
	import LoginLogoErrorSuccessIndicator from '../../components/login-logo-error-success-indicator.svelte';
	import Qrcode from '$lib/components/qrcode/qrcode.svelte';
	import { mode } from 'mode-watcher';

	let { data } = $props();

	const webauthnService = new WebAuthnService();

	let requesterIp = $state('');
	let requesterUserAgent = $state('');
	let expiresAt: string | null = $state(null);
	let authenticatorUrl = $state('');
	let exchangeToken = $state('');
	let isLoading = $state(true);
	let isPolling = $state(false);
	let timeSecondsLeft = $state(-1);
	let success = $state(false);
	let error: string | undefined = $state();
	let backHref = $state('/login/alternative');
	let pollHandle: ReturnType<typeof setInterval> | null = null;

	// If the previous page is a Pocket ID page, go back there instead of the generic alternative login page
	afterNavigate((e) => {
		if (e.from?.url.pathname) {
			backHref = e.from.url.pathname + e.from.url.search;
		}
	});

	function stopPolling() {
		if (pollHandle) {
			clearInterval(pollHandle);
			pollHandle = null;
			isPolling = false;
		}
	}

	async function pollStatus() {
		if (!exchangeToken) return;

		try {
			const status = await webauthnService.getCrossDeviceLoginStatus(exchangeToken);
			timeSecondsLeft = Math.max(
				0,
				Math.floor((new Date(expiresAt!).getTime() - Date.now()) / 1000)
			);
			if (status.status === 'completed' && status.user) {
				await userStore.setUser(status.user);
				success = true;
				stopPolling();

				try {
					goto(data.redirect);
				} catch (e) {
					error = m.invalid_redirect_url();
				}
			}
		} catch (e) {
			error = getWebauthnErrorMessage(e);
			stopPolling();
		}
	}

	onMount(async () => {
		isLoading = true;
		try {
			const res = await webauthnService.createCrossDeviceLogin();
			requesterIp = res.requesterIp;
			requesterUserAgent = res.requesterUserAgent;
			expiresAt = res.expiresAt;
			authenticatorUrl = res.authenticatorUrl;
			exchangeToken = res.exchangeToken;

			isPolling = true;
			pollHandle = setInterval(pollStatus, 3000);
		} catch (e) {
			error = getWebauthnErrorMessage(e);
		} finally {
			isLoading = false;
		}
	});

	onDestroy(() => {
		stopPolling();
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
		<p class="text-muted-foreground mt-2">
			Authentication complete. You can return to your requester device.
		</p>
		<div class="mt-6 flex flex-col items-center gap-2 text-sm text-muted-foreground">
			<p>Requester IP: {requesterIp}</p>
			<p>Requester browser: {requesterUserAgent}</p>
		</div>
	{:else if error}
		<h1 class="font-playfair mt-5 text-4xl font-bold">{m.sign_in()}</h1>
		<p class="text-muted-foreground mt-2">{error}</p>
		<div class="mt-8 flex justify-between gap-2">
			<Button variant="secondary" class="flex-1" href={backHref}>{m.go_back()}</Button>
		</div>
	{:else}
		<h1 class="font-playfair mt-5 text-4xl font-bold">{m.sign_in()}</h1>
		<p class="text-muted-foreground mt-2">Scan this QR code with your phone or use a login code.</p>

		{#if isLoading}
			<p class="mt-6 text-center text-sm text-muted-foreground">Generating QR code…</p>
		{:else}
			<div class="mt-6 flex flex-col items-center gap-4 rounded-lg border p-4">
				{#if authenticatorUrl}
					<Qrcode
						class="bg-background"
						value={authenticatorUrl}
						size={200}
						color={mode.current === 'dark' ? '#FFFFFF' : '#000000'}
						backgroundColor={mode.current === 'dark' ? '#000000' : '#FFFFFF'}
					/>
					<div class="text-xs text-muted-foreground space-y-1 text-center">
						<p>Authenticator URL: {authenticatorUrl}</p>
						{#if expiresAt}
							<p>Expires in {timeSecondsLeft}s</p>
						{/if}
					</div>
				{/if}
			</div>

			{#if isPolling}
				<p class="mt-6 text-center text-sm text-muted-foreground">
					Waiting for approval on your phone…
				</p>
			{/if}
		{/if}

		<div class="mt-8 flex justify-between gap-2">
			<Button variant="secondary" class="flex-1" href={backHref}>{m.go_back()}</Button>
		</div>
	{/if}
</SignInWrapper>
