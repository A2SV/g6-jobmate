package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tsigemariamzewdu/JobMate-backend/delivery/controllers"
	"github.com/tsigemariamzewdu/JobMate-backend/delivery/routes"
	// "github.com/tsigemariamzewdu/JobMate-backend/domain"
	groqClient "github.com/tsigemariamzewdu/JobMate-backend/infrastructure/ai"
	"github.com/tsigemariamzewdu/JobMate-backend/infrastructure/ai_service"
	authinfra "github.com/tsigemariamzewdu/JobMate-backend/infrastructure/auth"
	config "github.com/tsigemariamzewdu/JobMate-backend/infrastructure/config"
	// infrastructure "github.com/tsigemariamzewdu/JobMate-backend/infrastructure/config"
	infrastructure2 "github.com/tsigemariamzewdu/JobMate-backend/infrastructure/email"

	mongoclient "github.com/tsigemariamzewdu/JobMate-backend/infrastructure/db/mongo"
	utils "github.com/tsigemariamzewdu/JobMate-backend/infrastructure/util"
	"github.com/tsigemariamzewdu/JobMate-backend/repositories"
	"github.com/tsigemariamzewdu/JobMate-backend/usecases"
	// "google.golang.org/grpc/internal/resolver"
)

func main() {

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to MongoDB
	client := mongoclient.NewMongoClient()
	db := client.Database(cfg.DBName)
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB client: %v", err)
		}
	}()

	// Initialize repositories
	otpRepo := repositories.NewOTPRepository(db)
	authRepo := repositories.NewAuthRepository(db)
	userRepo := repositories.NewUserRepository(db)
	conversationRepo := repositories.NewConversationRepository(db)
	cvRepo := repositories.NewCVRepository(db)
	feedbackRepo := repositories.NewFeedbackRepository(db)
	skiilGapRepo := repositories.NewSkillGapRepository(db)

	providersConfigs, err := config.BuildProviderConfigs()
	if err != nil {
		log.Fatal("error: ", err)
	}

	// Initialize services
	phoneValidator := &authinfra.PhoneValidatorImpl{}
	emailService := infrastructure2.NewSMTPService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.EmailFrom)
	otpSender, err := authinfra.NewOTPSenderFromEnv(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize OTP sender: %v", err)
	}
	otpSenderTyped := otpSender
	jwtService := authinfra.NewJWTService(cfg.JWTSecretKey, fmt.Sprint(cfg.JWTExpirationMinutes))
	passwordService := authinfra.NewPasswordService()
	authMiddleware := authinfra.NewAuthMiddleware(jwtService)
	oauthService, err := authinfra.NewOAuth2Service(providersConfigs)
	aiService := ai_service.NewGeminiAISuggestionService("gemini-1.5-flash", cfg.AIApiKey)

	textExtractor := utils.NewFileTextExtractor()

	if err != nil {
		log.Fatalf("Failed to initialize OAuth2 service: %v", err)
	}

	// Initialize AI client
	groqClient := groqClient.NewGroqClient(cfg)

	// Initialize use case
	otpUsecase := usecases.NewOTPUsecase(otpRepo, phoneValidator, otpSenderTyped,emailService)
	authUsecase := usecases.NewAuthUsecase(authRepo, passwordService, jwtService, cfg.BaseURL,otpRepo, time.Second*10)
	userUsecase := usecases.NewUserUsecase(userRepo, time.Second*1000)
	chatUsecase := usecases.NewChatUsecase(conversationRepo, groqClient, cfg)
	cvUsecase := usecases.NewCVUsecase(cvRepo, feedbackRepo, skiilGapRepo, aiService, textExtractor, time.Second*15)

	// Initialize controllers
	otpController := controllers.NewOtpController(otpUsecase)
	authController := controllers.NewAuthController(authUsecase)
	userController := controllers.NewUserController(userUsecase)
	oauthController := controllers.NewOAuth2Controller(oauthService, authUsecase)
	chatController := controllers.NewChatController(chatUsecase)
	cvController := controllers.NewCVController(cvUsecase)

	// Setup router (add more controllers as you add features)
	router := routes.SetupRouter(authMiddleware, userController, authController, otpController, oauthController, chatController, cvController)

	// Get port from config or environment variable
	port := cfg.AppPort
	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
	}

	// Start server
	log.Printf("Server starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
