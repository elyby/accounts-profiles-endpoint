package http

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var timeNow = time.Now

var emptyTextures = []byte("{}")

type AccountsRepository interface {
	FindUsernameByUuid(ctx context.Context, uuid string) (string, error)
	// Should return uuid, correctly cased username and an error
	FindUuidByUsername(ctx context.Context, username string) (string, string, error)
}

type TexturesProvider interface {
	GetTexturesByUsername(ctx context.Context, username string) ([]byte, error)
}

// SignerService uses context because in the future we may separate this logic as an external microservice
type SignerService interface {
	Sign(ctx context.Context, data []byte) ([]byte, error)
	GetPublicKey(ctx context.Context, format string) ([]byte, error)
}

type MojangApi struct {
	AccountsRepository
	TexturesProvider
	SignerService
}

func NewMojangApi(
	accountsRepository AccountsRepository,
	texturesProvider TexturesProvider,
	signerService SignerService,
) MojangApi {
	return MojangApi{
		AccountsRepository: accountsRepository,
		TexturesProvider:   texturesProvider,
		SignerService:      signerService,
	}
}

func (s *MojangApi) DefineRoutes(r gin.IRouter) {
	r.GET("/api/minecraft/session/profile/:uuid", s.getProfileByUuidHandler)
	r.GET("/api/mojang/profiles/:username", s.getUuidByUsernameHandler)
}

func (s *MojangApi) getProfileByUuidHandler(c *gin.Context) {
	uuid, err := formatUuid(c.Param("uuid"))
	if err != nil {
		c.Status(http.StatusNoContent)
		return
	}

	username, err := s.AccountsRepository.FindUsernameByUuid(c.Request.Context(), uuid)
	if err != nil {
		c.Error(fmt.Errorf("unable to retrieve account information: %w", err))
		return
	}

	if username == "" {
		c.Status(http.StatusNoContent)
		return
	}

	textures, err := s.TexturesProvider.GetTexturesByUsername(c.Request.Context(), username)
	if err != nil {
		c.Error(fmt.Errorf("unable to retrieve textures information: %w", err))
		return
	}

	if textures == nil {
		textures = emptyTextures
	}

	serializedProfile, err := s.createProfileResponse(uuid, username, textures, c.Query("unsigned") == "false")
	if err != nil {
		c.Error(fmt.Errorf("unable to create a profile response: %w", err))
		return
	}

	c.Data(http.StatusOK, "application/json", serializedProfile)
}

func (s *MojangApi) getUuidByUsernameHandler(c *gin.Context) {
	uuid, username, err := s.AccountsRepository.FindUuidByUsername(c.Request.Context(), c.Param("username"))
	if err != nil {
		c.Error(fmt.Errorf("unable to retrieve user's uuid and correctly cased username: %w", err))
		return
	}

	if uuid == "" {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":   strings.ReplaceAll(uuid, "-", ""),
		"name": username,
	})
}

func (s *MojangApi) createProfileResponse(
	uuid string,
	username string,
	texturesJson []byte,
	sign bool,
) ([]byte, error) {
	uuidWithoutDashes := strings.ReplaceAll(uuid, "-", "")

	texturesPropValueJson := fmt.Appendf(
		nil,
		`{"timestamp":%d,"profileId":%q,"profileName":%q,"textures":%s}`,
		timeNow().UnixMilli(),
		uuidWithoutDashes,
		username,
		texturesJson,
	)

	encodedTexturesBuf := make([]byte, base64.StdEncoding.EncodedLen(len(texturesPropValueJson)))
	base64.StdEncoding.Encode(encodedTexturesBuf, texturesPropValueJson)

	result := fmt.Appendf(
		nil,
		`{"id":%q,"name":%q,"properties":[{"name":"textures","value":%q`,
		uuidWithoutDashes,
		username,
		encodedTexturesBuf,
	)

	if sign {
		signature, err := s.SignerService.Sign(context.Background(), encodedTexturesBuf)
		if err != nil {
			return nil, fmt.Errorf("unable to sign textures: %w", err)
		}

		encodedSignatureBuf := make([]byte, base64.StdEncoding.EncodedLen(len(signature)))
		base64.StdEncoding.Encode(encodedSignatureBuf, signature)

		result = fmt.Appendf(result, `,"signature":%q`, encodedSignatureBuf)
	}

	result = fmt.Appendf(result, `},{"name":"ely","value":"but why are you asking?"}]}`)

	return result, nil
}

var invalidUuid = errors.New("invalid uuid")

func formatUuid(input string) (string, error) {
	uuid := strings.ReplaceAll(input, "-", "")
	if len(uuid) != 32 {
		return "", invalidUuid
	}

	return fmt.Sprintf("%s-%s-%s-%s-%s", uuid[0:8], uuid[8:12], uuid[12:16], uuid[16:20], uuid[20:]), nil
}
