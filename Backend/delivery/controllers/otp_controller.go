package controllers

import (
	"context"
	"github.com/tsigemariamzewdu/JobMate-backend/delivery/dto"
	"github.com/tsigemariamzewdu/JobMate-backend/usecases"
	"github.com/tsigemariamzewdu/JobMate-backend/domain/models"

	"net/http"
	"fmt"

	"github.com/gin-gonic/gin"
)

type OtpController struct {
    AuthUsecase *usecases.OTPUsecase
}

func NewOtpController(authUsecase *usecases.OTPUsecase) *OtpController {
    return &OtpController{AuthUsecase: authUsecase}
}

// POST /auth/request-otp
func (c *OtpController) RequestOTP(ctx *gin.Context) {
    var req dto.OTPRequestDTO
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, dto.OTPResponseDTO{Message: "Invalid request"})
        return
    }
    // Get requestor IP
    ip := ctx.ClientIP()
    otpReq := dtoToDomainOTPRequest(req, ip)
    if err := c.AuthUsecase.RequestOTP(context.Background(), &otpReq); err != nil {
        // log it but don’t expose to client
        fmt.Printf("failed to send OTP: %v\n", err)
    }
    ctx.JSON(http.StatusOK, dto.OTPResponseDTO{Message: "If this email exists, a code was sent"})

}

func dtoToDomainOTPRequest(req dto.OTPRequestDTO, ip string) models.OTPRequest {
    return models.OTPRequest{
        Email: req.Email,
        RequestorIP: ip,
    }
}
