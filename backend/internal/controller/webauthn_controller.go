package controller

import (
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/pocket-id/pocket-id/backend/internal/common"
	"github.com/pocket-id/pocket-id/backend/internal/dto"
	"github.com/pocket-id/pocket-id/backend/internal/middleware"
	"github.com/pocket-id/pocket-id/backend/internal/utils/cookie"

	"github.com/gin-gonic/gin"
	"github.com/pocket-id/pocket-id/backend/internal/service"
	"golang.org/x/time/rate"
)

func NewWebauthnController(group *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware, rateLimitMiddleware *middleware.RateLimitMiddleware, webauthnService *service.WebAuthnService, appConfigService *service.AppConfigService) {
	wc := &WebauthnController{webAuthnService: webauthnService, appConfigService: appConfigService}
	group.GET("/webauthn/register/start", authMiddleware.WithAdminNotRequired().Add(), wc.beginRegistrationHandler)
	group.POST("/webauthn/register/finish", authMiddleware.WithAdminNotRequired().Add(), wc.verifyRegistrationHandler)

	group.GET("/webauthn/login/start", wc.beginLoginHandler)
	group.POST("/webauthn/login/finish", rateLimitMiddleware.Add(rate.Every(10*time.Second), 5), wc.verifyLoginHandler)

	group.POST("/webauthn/cross-device/start", wc.createCrossDeviceLoginHandler)
	group.GET("/webauthn/cross-device/login/start", wc.beginCrossDeviceLoginHandler)
	group.POST("/webauthn/cross-device/login/finish", rateLimitMiddleware.Add(rate.Every(10*time.Second), 5), wc.finishCrossDeviceLoginHandler)
	group.GET("/webauthn/cross-device/status", wc.crossDeviceLoginStatusHandler)

	group.POST("/webauthn/logout", authMiddleware.WithAdminNotRequired().Add(), wc.logoutHandler)

	group.POST("/webauthn/reauthenticate", authMiddleware.WithAdminNotRequired().Add(), rateLimitMiddleware.Add(rate.Every(10*time.Second), 5), wc.reauthenticateHandler)

	group.GET("/webauthn/credentials", authMiddleware.WithAdminNotRequired().Add(), wc.listCredentialsHandler)
	group.PATCH("/webauthn/credentials/:id", authMiddleware.WithAdminNotRequired().Add(), wc.updateCredentialHandler)
	group.DELETE("/webauthn/credentials/:id", authMiddleware.WithAdminNotRequired().Add(), wc.deleteCredentialHandler)
}

type WebauthnController struct {
	webAuthnService  *service.WebAuthnService
	appConfigService *service.AppConfigService
}

func (wc *WebauthnController) beginRegistrationHandler(c *gin.Context) {
	userID := c.GetString("userID")
	options, err := wc.webAuthnService.BeginRegistration(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	cookie.AddSessionIdCookie(c, int(options.Timeout.Seconds()), options.SessionID)
	c.JSON(http.StatusOK, options.Response)
}

func (wc *WebauthnController) verifyRegistrationHandler(c *gin.Context) {
	sessionID, err := c.Cookie(cookie.SessionIdCookieName)
	if err != nil {
		_ = c.Error(&common.MissingSessionIdError{})
		return
	}

	userID := c.GetString("userID")
	credential, err := wc.webAuthnService.VerifyRegistration(c.Request.Context(), sessionID, userID, c.Request, c.ClientIP())
	if err != nil {
		_ = c.Error(err)
		return
	}

	var credentialDto dto.WebauthnCredentialDto
	if err := dto.MapStruct(credential, &credentialDto); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, credentialDto)
}

func (wc *WebauthnController) beginLoginHandler(c *gin.Context) {
	options, err := wc.webAuthnService.BeginLogin(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	cookie.AddSessionIdCookie(c, int(options.Timeout.Seconds()), options.SessionID)
	c.JSON(http.StatusOK, options.Response)
}

func (wc *WebauthnController) createCrossDeviceLoginHandler(c *gin.Context) {
	crossDeviceLogin, exchangeToken, err := wc.webAuthnService.CreateCrossDeviceLogin(c.Request.Context(), c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		_ = c.Error(err)
		return
	}

	authenticatorURL := common.EnvConfig.AppURL + "/login/cross-device?code=" + crossDeviceLogin.Code
	resp := dto.WebauthnCrossDeviceStartResponseDto{
		Code:          crossDeviceLogin.Code,
		AuthenticatorURL:         authenticatorURL,
		ExchangeToken: exchangeToken,
		ExpiresAt:     crossDeviceLogin.ExpiresAt,
		RequesterIP:   crossDeviceLogin.RequesterIP,
		RequesterUserAgent: crossDeviceLogin.RequesterUserAgent,
	}

	c.JSON(http.StatusOK, resp)
}

func (wc *WebauthnController) beginCrossDeviceLoginHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		_ = c.Error(&common.CrossDeviceLoginInvalidError{})
		return
	}

	qrLogin, options, err := wc.webAuthnService.BeginCrossDeviceLogin(c.Request.Context(), code)
	if err != nil {
		_ = c.Error(err)
		return
	}

	cookie.AddSessionIdCookie(c, int(options.Timeout.Seconds()), options.SessionID)

	resp := dto.WebauthnCrossDeviceLoginStartResponseDto{
		Response:       options.Response,
		SessionID:      options.SessionID,
		Timeout:        options.Timeout,
		RequesterIP:    qrLogin.RequesterIP,
		RequesterUserAgent: qrLogin.RequesterUserAgent,
		ExpiresAt:      qrLogin.ExpiresAt,
	}

	c.JSON(http.StatusOK, resp)
}

func (wc *WebauthnController) finishCrossDeviceLoginHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		_ = c.Error(&common.CrossDeviceLoginInvalidError{})
		return
	}

	sessionID, err := c.Cookie(cookie.SessionIdCookieName)
	if err != nil {
		_ = c.Error(&common.MissingSessionIdError{})
		return
	}

	credentialAssertionData, err := protocol.ParseCredentialRequestResponseBody(c.Request.Body)
	if err != nil {
		_ = c.Error(err)
		return
	}

	user, err := wc.webAuthnService.CompleteCrossDeviceLogin(c.Request.Context(), code, sessionID, credentialAssertionData, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		_ = c.Error(err)
		return
	}

	var userDto dto.UserDto
	if err := dto.MapStruct(user, &userDto); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, userDto)
}

func (wc *WebauthnController) verifyLoginHandler(c *gin.Context) {
	sessionID, err := c.Cookie(cookie.SessionIdCookieName)
	if err != nil {
		_ = c.Error(&common.MissingSessionIdError{})
		return
	}

	credentialAssertionData, err := protocol.ParseCredentialRequestResponseBody(c.Request.Body)
	if err != nil {
		_ = c.Error(err)
		return
	}

	user, token, err := wc.webAuthnService.VerifyLogin(c.Request.Context(), sessionID, credentialAssertionData, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		_ = c.Error(err)
		return
	}

	var userDto dto.UserDto
	if err := dto.MapStruct(user, &userDto); err != nil {
		_ = c.Error(err)
		return
	}

	maxAge := int(wc.appConfigService.GetDbConfig().SessionDuration.AsDurationMinutes().Seconds())
	cookie.AddAccessTokenCookie(c, maxAge, token)

	c.JSON(http.StatusOK, userDto)
}

func (wc *WebauthnController) crossDeviceLoginStatusHandler(c *gin.Context) {
	exchangeToken := c.Query("exchangeToken")
	if exchangeToken == "" {
		_ = c.Error(&common.CrossDeviceLoginInvalidError{})
		return
	}

	crossDeviceLogin, user, token, err := wc.webAuthnService.PollCrossDeviceLoginStatus(c.Request.Context(), exchangeToken)
	if err != nil {
		_ = c.Error(err)
		return
	}

	resp := dto.WebauthnCrossDeviceStatusResponseDto{
		Status:    "pending",
		ExpiresAt: crossDeviceLogin.ExpiresAt,
	}

	if user != nil {
		maxAge := int(wc.appConfigService.GetDbConfig().SessionDuration.AsDurationMinutes().Seconds())
		cookie.AddAccessTokenCookie(c, maxAge, token)

		var userDto dto.UserDto
		if err := dto.MapStruct(user, &userDto); err != nil {
			_ = c.Error(err)
			return
		}

		resp.Status = "completed"
		resp.User = &userDto
	}

	c.JSON(http.StatusOK, resp)
}

func (wc *WebauthnController) listCredentialsHandler(c *gin.Context) {
	userID := c.GetString("userID")
	credentials, err := wc.webAuthnService.ListCredentials(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var credentialDtos []dto.WebauthnCredentialDto
	if err := dto.MapStructList(credentials, &credentialDtos); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, credentialDtos)
}

func (wc *WebauthnController) deleteCredentialHandler(c *gin.Context) {
	userID := c.GetString("userID")
	credentialID := c.Param("id")
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	err := wc.webAuthnService.DeleteCredential(c.Request.Context(), userID, credentialID, clientIP, userAgent)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (wc *WebauthnController) updateCredentialHandler(c *gin.Context) {
	userID := c.GetString("userID")
	credentialID := c.Param("id")

	var input dto.WebauthnCredentialUpdateDto
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(err)
		return
	}

	credential, err := wc.webAuthnService.UpdateCredential(c.Request.Context(), userID, credentialID, input.Name)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var credentialDto dto.WebauthnCredentialDto
	if err := dto.MapStruct(credential, &credentialDto); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, credentialDto)
}

func (wc *WebauthnController) logoutHandler(c *gin.Context) {
	cookie.AddAccessTokenCookie(c, 0, "")
	c.Status(http.StatusNoContent)
}

func (wc *WebauthnController) reauthenticateHandler(c *gin.Context) {
	sessionID, err := c.Cookie(cookie.SessionIdCookieName)
	if err != nil {
		_ = c.Error(&common.MissingSessionIdError{})
		return
	}

	var token string

	// Try to create a reauthentication token with WebAuthn
	credentialAssertionData, err := protocol.ParseCredentialRequestResponseBody(c.Request.Body)
	if err == nil {
		token, err = wc.webAuthnService.CreateReauthenticationTokenWithWebauthn(c.Request.Context(), sessionID, credentialAssertionData)
		if err != nil {
			_ = c.Error(err)
			return
		}
	} else {
		// If WebAuthn fails, try to create a reauthentication token with the access token
		accessToken, _ := c.Cookie(cookie.AccessTokenCookieName)
		token, err = wc.webAuthnService.CreateReauthenticationTokenWithAccessToken(c.Request.Context(), accessToken)
		if err != nil {
			_ = c.Error(err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"reauthenticationToken": token})
}
