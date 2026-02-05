package routes

import (
	"wallet-point/internal/audit"
	"wallet-point/internal/auth"
	"wallet-point/internal/marketplace"
	"wallet-point/internal/mission"
	"wallet-point/internal/transfer"
	"wallet-point/internal/user"
	"wallet-point/internal/wallet"
	"wallet-point/middleware"
	"wallet-point/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	// Swagger imports
	_ "wallet-point/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, allowedOrigins string, jwtExpiry int) {
	// Apply global middleware
	r.Use(middleware.CORS(allowedOrigins))
	r.Use(middleware.Logger())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.IPBasedRateLimiter())

	r.Static("/uploads", "../../public/uploads")

	api := r.Group("/api/v1")

	// Global Upload Endpoint
	api.POST("/upload", middleware.AuthMiddleware(), utils.HandleFileUpload)

	// Initialize repositories
	authRepo := auth.NewAuthRepository(db)
	userRepo := user.NewUserRepository(db)
	walletRepo := wallet.NewWalletRepository(db)
	marketplaceRepo := marketplace.NewMarketplaceRepository(db)
	auditRepo := audit.NewAuditRepository(db)
	missionRepo := mission.NewMissionRepository(db)

	// Initialize services
	authService := auth.NewAuthService(authRepo, jwtExpiry)
	userService := user.NewUserService(userRepo)
	walletService := wallet.NewWalletService(walletRepo, db)
	walletService.SetAuthService(authService) // Inject for PIN verification

	marketplaceService := marketplace.NewMarketplaceService(marketplaceRepo, walletService, authService, db)
	auditService := audit.NewAuditService(auditRepo)
	missionService := mission.NewMissionService(missionRepo, walletService, db)
	transferService := transfer.NewService(walletRepo, walletService, authService, db)

	// Initialize handlers
	authHandler := auth.NewAuthHandler(authService, auditService)
	userHandler := user.NewUserHandler(userService, auditService)
	walletHandler := wallet.NewWalletHandler(walletService, auditService)
	marketplaceHandler := marketplace.NewMarketplaceHandler(marketplaceService, auditService)
	auditHandler := audit.NewAuditHandler(auditService)
	missionHandler := mission.NewMissionHandler(missionService, auditService)
	transferHandler := transfer.NewHandler(transferService, auditService)

	// ========================================
	// PUBLIC ROUTES
	// ========================================
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", middleware.AuthRateLimiter(), authHandler.Login)
		authGroup.POST("/register", middleware.AuthRateLimiter(), authHandler.PublicRegister)
		authGroup.GET("/me", middleware.AuthMiddleware(), authHandler.Me)
		authGroup.PUT("/profile", middleware.AuthMiddleware(), authHandler.UpdateProfile)
		authGroup.PUT("/password", middleware.AuthMiddleware(), authHandler.UpdatePassword)
		authGroup.PUT("/pin", middleware.AuthMiddleware(), authHandler.UpdatePin)
	}

	// ========================================
	// ADMIN ROUTES
	// ========================================
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.RoleMiddleware("admin"))
	{
		// User Management
		adminGroup.POST("/users", authHandler.Register)
		adminGroup.GET("/users", userHandler.GetAll)
		adminGroup.GET("/users/:id", userHandler.GetByID)
		adminGroup.PUT("/users/:id", userHandler.Update)
		adminGroup.DELETE("/users/:id", userHandler.Deactivate)
		adminGroup.PUT("/users/:id/password", userHandler.ChangePassword)

		// Wallet Management
		adminGroup.GET("/wallets", walletHandler.GetAllWallets)
		adminGroup.GET("/wallets/:id", walletHandler.GetWalletByID)
		adminGroup.GET("/wallets/:id/transactions", walletHandler.GetWalletTransactions)
		adminGroup.POST("/wallet/adjustment", walletHandler.AdjustPoints)
		adminGroup.POST("/wallet/reset", walletHandler.ResetWallet)

		// Transaction Monitoring
		adminGroup.GET("/transactions", walletHandler.GetAllTransactions)
		adminGroup.GET("/transfers", transferHandler.GetAllTransfers)

		// Marketplace Management
		adminGroup.GET("/marketplace/transactions", marketplaceHandler.GetTransactions)
		adminGroup.GET("/products", marketplaceHandler.GetAll)
		adminGroup.POST("/products", marketplaceHandler.Create)
		adminGroup.GET("/products/:id", marketplaceHandler.GetByID)
		adminGroup.PUT("/products/:id", marketplaceHandler.Update)
		adminGroup.DELETE("/products/:id", marketplaceHandler.Delete)

		// Audit Logs
		adminGroup.GET("/audit-logs", auditHandler.GetAll)

		// Admin Dashboard Stats
		adminGroup.GET("/stats", walletHandler.GetAdminStats)
	}

	// ========================================
	// DOSEN ROUTES
	// ========================================
	dosenGroup := api.Group("/dosen")
	dosenGroup.Use(middleware.AuthMiddleware())
	dosenGroup.Use(middleware.RoleMiddleware("dosen", "admin"))
	{
		// Mission & Task Management
		dosenGroup.POST("/missions", missionHandler.CreateMission)
		dosenGroup.PUT("/missions/:id", missionHandler.UpdateMission)
		dosenGroup.DELETE("/missions/:id", missionHandler.DeleteMission)
		dosenGroup.GET("/missions", missionHandler.GetAllMissions)
		dosenGroup.GET("/missions/:id", missionHandler.GetMissionByID)

		// Submission Validation
		dosenGroup.GET("/submissions", missionHandler.GetAllSubmissions)
		dosenGroup.POST("/submissions/:id/review", missionHandler.ReviewSubmission)

		// Dashboard Stats
		dosenGroup.GET("/stats", missionHandler.GetDosenStats)

		// Monitoring & Manual Rewards
		dosenGroup.GET("/students", userHandler.GetAll)
		dosenGroup.POST("/reward", walletHandler.AdjustPoints)
	}

	// ========================================
	// MAHASISWA ROUTES
	// ========================================
	mahasiswaGroup := api.Group("/mahasiswa")
	mahasiswaGroup.Use(middleware.AuthMiddleware())
	mahasiswaGroup.Use(middleware.RoleMiddleware("mahasiswa"))
	{
		// Mission & Task Submission
		mahasiswaGroup.GET("/missions", missionHandler.GetAllMissions)
		mahasiswaGroup.GET("/missions/:id", missionHandler.GetMissionByID)
		mahasiswaGroup.POST("/missions/submit", missionHandler.SubmitMission)
		mahasiswaGroup.GET("/submissions", missionHandler.GetAllSubmissions)

		// Transfer Points
		mahasiswaGroup.POST("/transfer", transferHandler.CreateTransfer)
		mahasiswaGroup.GET("/transfer/history", transferHandler.GetMyTransfers)
		mahasiswaGroup.GET("/transfer/recipient/:id", transferHandler.GetRecipientInfo)
		mahasiswaGroup.GET("/users/lookup", userHandler.LookupUser)

		// Marketplace & Cart
		mahasiswaGroup.GET("/marketplace/products", marketplaceHandler.GetAll)
		mahasiswaGroup.GET("/marketplace/products/:id", marketplaceHandler.GetByID)
		mahasiswaGroup.POST("/marketplace/purchase", marketplaceHandler.Purchase)
		mahasiswaGroup.GET("/marketplace/cart", marketplaceHandler.GetCart)
		mahasiswaGroup.POST("/marketplace/cart", marketplaceHandler.AddToCart)
		mahasiswaGroup.PUT("/marketplace/cart/:id", marketplaceHandler.UpdateCartItem)
		mahasiswaGroup.DELETE("/marketplace/cart/:id", marketplaceHandler.RemoveFromCart)
		mahasiswaGroup.POST("/marketplace/cart/checkout", marketplaceHandler.Checkout)

		// Gamification
		mahasiswaGroup.GET("/leaderboard", walletHandler.GetLeaderboard)

		// Personal Wallet
		mahasiswaGroup.GET("/wallet", walletHandler.GetMyWallet)
		mahasiswaGroup.GET("/transactions", walletHandler.GetMyTransactions)
		mahasiswaGroup.POST("/payment/token", walletHandler.GeneratePaymentToken)
		mahasiswaGroup.POST("/payment/execute", walletHandler.ExecuteStudentPayment)
	}
	// Global QR Status Check
	api.GET("/payment/status/:token", walletHandler.CheckTokenStatus)
	api.GET("/missions/:id/leaderboard", missionHandler.GetQuizLeaderboard)

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Wallet Point API is running",
		})
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
