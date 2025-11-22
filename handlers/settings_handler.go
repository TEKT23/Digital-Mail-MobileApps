package handlers

import (
	"TugasAkhir/config"
	userdto "TugasAkhir/dto/users"
	"TugasAkhir/middleware"
	"TugasAkhir/models"
	"TugasAkhir/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetMyProfile(c *fiber.Ctx) error {
	claims, ok := middleware.GetJWTClaims(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var user models.User
	if err := config.DB.First(&user, claims.UserID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "user not found", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "profile retrieved", toUserSummary(user))

}

func UpdateMyProfile(c *fiber.Ctx) error {
	claims, ok := middleware.GetJWTClaims(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req userdto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid body", err.Error())
	}

	var user models.User
	if err := config.DB.First(&user, claims.UserID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "user not found", nil)
	}
	user.FirstName = strings.TrimSpace(req.FirstName)
	user.LastName = strings.TrimSpace(req.LastName)
	user.Jabatan = strings.TrimSpace(req.Jabatan)
	user.Atribut = strings.TrimSpace(req.Atribut)

	if err := config.DB.Save(&user).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to update profile", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "profile updated successfully", toUserSummary(user))
}

func ChangePassword(c *fiber.Ctx) error {
	claims, ok := middleware.GetJWTClaims(c)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req userdto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid body", err.Error())
	}

	if errs := req.Validate(); len(errs) > 0 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "validation error", errs)
	}

	var user models.User
	if err := config.DB.First(&user, claims.UserID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "user not found", nil)
	}

	if !utils.CheckPassword(user.PasswordHash, req.OldPassword) {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "password lama salah", nil)
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to hash password", err.Error())
	}

	user.PasswordHash = newHash
	if err := config.DB.Save(&user).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to update password", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "password updated successfully", nil)

}
