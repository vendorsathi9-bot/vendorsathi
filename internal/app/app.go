package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"vendorsaathi/internal/config"
	"vendorsaathi/internal/middleware"
	analyticshandler "vendorsaathi/internal/modules/analytics/handler"
	analyticsrepository "vendorsaathi/internal/modules/analytics/repository"
	analyticsservice "vendorsaathi/internal/modules/analytics/service"
	audithandler "vendorsaathi/internal/modules/audit/handler"
	auditrepository "vendorsaathi/internal/modules/audit/repository"
	auditservice "vendorsaathi/internal/modules/audit/service"
	authhandler "vendorsaathi/internal/modules/auth/handler"
	authrepository "vendorsaathi/internal/modules/auth/repository"
	authservice "vendorsaathi/internal/modules/auth/service"
	cartshandler "vendorsaathi/internal/modules/carts/handler"
	cartsrepository "vendorsaathi/internal/modules/carts/repository"
	cartsservice "vendorsaathi/internal/modules/carts/service"
	customershandler "vendorsaathi/internal/modules/customers/handler"
	customersrepository "vendorsaathi/internal/modules/customers/repository"
	customersservice "vendorsaathi/internal/modules/customers/service"
	deliveryhandler "vendorsaathi/internal/modules/delivery/handler"
	deliveryrepository "vendorsaathi/internal/modules/delivery/repository"
	deliveryservice "vendorsaathi/internal/modules/delivery/service"
	inventoryHandler "vendorsaathi/internal/modules/inventory/handler"
	inventoryRepository "vendorsaathi/internal/modules/inventory/repository"
	inventoryService "vendorsaathi/internal/modules/inventory/service"
	mobilityhandler "vendorsaathi/internal/modules/mobility/handler"
	mobilityrepository "vendorsaathi/internal/modules/mobility/repository"
	mobilityservice "vendorsaathi/internal/modules/mobility/service"
	notificationshandler "vendorsaathi/internal/modules/notifications/handler"
	notificationsrepository "vendorsaathi/internal/modules/notifications/repository"
	notificationsservice "vendorsaathi/internal/modules/notifications/service"
	ordershandler "vendorsaathi/internal/modules/orders/handler"
	ordersrepository "vendorsaathi/internal/modules/orders/repository"
	ordersservice "vendorsaathi/internal/modules/orders/service"
	paymentshandler "vendorsaathi/internal/modules/payments/handler"
	paymentsrepository "vendorsaathi/internal/modules/payments/repository"
	paymentsservice "vendorsaathi/internal/modules/payments/service"
	productsHandler "vendorsaathi/internal/modules/products/handler"
	productsRepository "vendorsaathi/internal/modules/products/repository"
	productsService "vendorsaathi/internal/modules/products/service"
	reviewshandler "vendorsaathi/internal/modules/reviews/handler"
	reviewsrepository "vendorsaathi/internal/modules/reviews/repository"
	reviewsservice "vendorsaathi/internal/modules/reviews/service"
	synchandler "vendorsaathi/internal/modules/sync/handler"
	syncrepository "vendorsaathi/internal/modules/sync/repository"
	syncservice "vendorsaathi/internal/modules/sync/service"
	udhaarhandler "vendorsaathi/internal/modules/udhaar/handler"
	udhaarrepository "vendorsaathi/internal/modules/udhaar/repository"
	udhaarservice "vendorsaathi/internal/modules/udhaar/service"
	usershandler "vendorsaathi/internal/modules/users/handler"
	usersrepository "vendorsaathi/internal/modules/users/repository"
	usersservice "vendorsaathi/internal/modules/users/service"
	vendorshandler "vendorsaathi/internal/modules/vendors/handler"
	vendorsrepository "vendorsaathi/internal/modules/vendors/repository"
	vendorsservice "vendorsaathi/internal/modules/vendors/service"
	"vendorsaathi/internal/platform/auth"
	"vendorsaathi/internal/platform/database"
	"vendorsaathi/internal/platform/logger"
)

type Application struct {
	cfg        *config.Config
	router     *gin.Engine
	server     *http.Server
	dbPool     *pgxpool.Pool
	jwtManager *auth.JWTManager
}

type Handlers struct {
	auth          *authhandler.AuthHandler
	users         *usershandler.UserHandler
	vendors       *vendorshandler.VendorHandler
	customers     *customershandler.CustomerHandler
	products      *productsHandler.ProductHandler
	inventory     *inventoryHandler.InventoryHandler
	carts         *cartshandler.CartHandler
	orders        *ordershandler.OrderHandler
	payments      *paymentshandler.PaymentHandler
	udhaar        *udhaarhandler.UdhaarHandler
	notifications *notificationshandler.NotificationHandler
	audit         *audithandler.AuditHandler
	sync          *synchandler.SyncHandler
	mobility      *mobilityhandler.MobilityHandler
	analytics     *analyticshandler.AnalyticsHandler
	reviews       *reviewshandler.ReviewHandler
	delivery      *deliveryhandler.DeliveryHandler
}

func New(cfg *config.Config) (*Application, error) {
	logger.Init(cfg.LogLevel, cfg.IsDevelopment())

	dbPool, err := database.NewPool(context.Background(), database.DefaultConfig(cfg.DatabaseURL))
	if err != nil {
		return nil, fmt.Errorf("database connection: %w", err)
	}

	jwtManager := auth.NewJWTManager(
		cfg.JWT.SigningKey,
		cfg.JWT.Issuer,
		cfg.JWT.Audience,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	authUserRepo := authrepository.NewUserRepository(dbPool)
	customersCustomerRepo := customersrepository.NewCustomerRepository(dbPool)
	txManager := database.NewTxManager(dbPool)
	authService := authservice.NewAuthService(authUserRepo, customersCustomerRepo, txManager, jwtManager, &database.RealClock{})
	authHandler := authhandler.NewAuthHandler(authService)

	usersUserRepo := usersrepository.NewUserRepository(dbPool)
	usersService := usersservice.NewUserService(usersUserRepo)
	usersHandler := usershandler.NewUserHandler(usersService)

	vendorsVendorRepo := vendorsrepository.NewVendorRepository(dbPool)
	vendorsService := vendorsservice.NewVendorService(vendorsVendorRepo)
	vendorsHandler := vendorshandler.NewVendorHandler(vendorsService)

	customersService := customersservice.NewCustomerService(customersCustomerRepo)
	customersHandler := customershandler.NewCustomerHandler(customersService)

	productsProductRepo := productsRepository.NewProductRepository(dbPool)
	productsService := productsService.NewProductService(productsProductRepo)
	productsHandler := productsHandler.NewProductHandler(productsService)

	inventoryInventoryRepo := inventoryRepository.NewInventoryRepository(dbPool)
	inventoryService := inventoryService.NewInventoryService(inventoryInventoryRepo)
	inventoryHandler := inventoryHandler.NewInventoryHandler(inventoryService)

	cartsRepo := cartsrepository.NewCartRepository(dbPool)
	cartsService := cartsservice.NewCartService(cartsRepo, productsProductRepo)
	cartsHandler := cartshandler.NewCartHandler(cartsService)

	ordersRepo := ordersrepository.NewOrderRepository(dbPool)
	ordersService := ordersservice.NewOrderService(dbPool, ordersRepo, cartsRepo, inventoryInventoryRepo, productsProductRepo)
	ordersHandler := ordershandler.NewOrderHandler(ordersService)

	paymentsRepo := paymentsrepository.NewPaymentRepository(dbPool)
	paymentsService := paymentsservice.NewPaymentService(paymentsRepo)
	paymentsHandler := paymentshandler.NewPaymentHandler(paymentsService, cfg.Payment)

	udhaarRepo := udhaarrepository.NewUdhaarRepository(dbPool)
	udhaarService := udhaarservice.NewUdhaarService(udhaarRepo)
	udhaarHandler := udhaarhandler.NewUdhaarHandler(udhaarService)

	notificationsRepo := notificationsrepository.NewNotificationRepository(dbPool)
	notificationsService := notificationsservice.NewNotificationService(notificationsRepo)
	notificationsHandler := notificationshandler.NewNotificationHandler(notificationsService)

	auditRepo := auditrepository.NewAuditRepository(dbPool)
	auditService := auditservice.NewAuditService(auditRepo)
	auditHandler := audithandler.NewAuditHandler(auditService)

	syncRepo := syncrepository.NewSyncRepository(dbPool)
	syncService := syncservice.NewSyncService(syncRepo, dbPool)
	syncHandler := synchandler.NewSyncHandler(syncService)

	mobilityRepo := mobilityrepository.NewMobilityRepository(dbPool)
	mobilityService := mobilityservice.NewMobilityService(mobilityRepo, cfg.Map)
	mobilityHandler := mobilityhandler.NewMobilityHandler(mobilityService, cfg.Map)

	analyticsRepo := analyticsrepository.NewAnalyticsRepository(dbPool)
	analyticsService := analyticsservice.NewAnalyticsService(analyticsRepo)
	analyticsHandler := analyticshandler.NewAnalyticsHandler(analyticsService)

	reviewsRepo := reviewsrepository.NewReviewRepository(dbPool)
	reviewsService := reviewsservice.NewReviewService(reviewsRepo)
	reviewsHandler := reviewshandler.NewReviewHandler(reviewsService)

	deliveryRepo := deliveryrepository.NewDeliveryRepository(dbPool)
	deliveryService := deliveryservice.NewDeliveryService(deliveryRepo)
	deliveryHandler := deliveryhandler.NewDeliveryHandler(deliveryService)

	handlers := Handlers{
		auth:          authHandler,
		users:         usersHandler,
		vendors:       vendorsHandler,
		customers:     customersHandler,
		products:      productsHandler,
		inventory:     inventoryHandler,
		carts:         cartsHandler,
		orders:        ordersHandler,
		payments:      paymentsHandler,
		udhaar:        udhaarHandler,
		notifications: notificationsHandler,
		audit:         auditHandler,
		sync:          syncHandler,
		mobility:      mobilityHandler,
		analytics:     analyticsHandler,
		reviews:       reviewsHandler,
		delivery:      deliveryHandler,
	}

	router := setupRouter(cfg, jwtManager, handlers, dbPool)

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  cfg.RequestTimeout,
		WriteTimeout: cfg.RequestTimeout,
		IdleTimeout:  120 * time.Second,
	}

	return &Application{
		cfg:        cfg,
		router:     router,
		server:     server,
		dbPool:     dbPool,
		jwtManager: jwtManager,
	}, nil
}

func setupRouter(cfg *config.Config, jwtManager *auth.JWTManager, h Handlers, dbPool *pgxpool.Pool) *gin.Engine {
	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.CORSMiddleware(cfg.CORSOrigins))

	if cfg.IsDevelopment() {
		router.Use(gin.Logger())
	}

	// Health endpoints
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := dbPool.Ping(ctx); err != nil {
			c.JSON(503, gin.H{"status": "not ready", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ready"})
	})

	// Public discovery routes
	router.GET("/api/v1/mobility/nearby-vendors", h.mobility.FindNearbyVendors)
	router.GET("/api/v1/mobility/route", h.mobility.GetRoute)
	router.GET("/api/v1/mobility/config", h.mobility.GetMapConfig)
	router.GET("/api/v1/mobility/tiles/:z/:x/:y", h.mobility.ProxyTile)
	router.GET("/api/v1/mobility/geocode", h.mobility.Geocode)
	router.GET("/api/v1/mobility/reverse-geocode", h.mobility.ReverseGeocode)

	// Public read operations (browsing vendors, products, categories)
	public := router.Group("/api/v1")
	public.Use(middleware.OptionalAuthMiddleware(jwtManager))
	{
		public.GET("/products", h.products.GetMyProducts)
		public.GET("/products/:id", h.products.GetProductByID)
		public.GET("/products/vendor/:vendorId", h.products.GetVendorProducts)
		public.GET("/categories", h.products.GetCategories)
		public.GET("/categories/:id", h.products.GetCategoryByID)
		public.GET("/categories/:id/products", h.products.GetProductsByCategory)
		public.GET("/vendors/:id", h.vendors.GetProfileByID)
		public.GET("/vendors/:id/detail", h.vendors.GetVendorDetail)
	}

	// Auth routes
	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/register", h.auth.Register)
		authGroup.POST("/login", h.auth.Login)
		authGroup.POST("/refresh", h.auth.Refresh)
		authGroup.POST("/logout", h.auth.Logout)
	}

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(jwtManager))
	{
		protected.GET("/me", h.auth.Me)

		// Users routes
		usersGroup := protected.Group("/users")
		{
			usersGroup.GET("/me", h.users.Me)
			usersGroup.PATCH("/me/profile", h.users.UpdateProfile)
			usersGroup.POST("/change-password", h.users.ChangePassword)
			usersGroup.GET("/:id", h.users.GetUser)
			usersGroup.GET("", h.users.ListUsers)
			usersGroup.PATCH("/:id/status", h.users.UpdateStatus)
			usersGroup.POST("/:id/roles", h.users.AssignRole)
			usersGroup.DELETE("/:id/roles", h.users.RevokeRole)
		}

		// Vendors routes
		vendorsGroup := protected.Group("/vendors")
		{
			vendorsGroup.GET("", h.vendors.ListVendors)
			vendorsGroup.POST("/profile", h.vendors.CreateProfile)
			vendorsGroup.GET("/profile/me", h.vendors.GetMyProfile)
			vendorsGroup.PATCH("/profile/me", h.vendors.UpdateMyProfile)
			vendorsGroup.PATCH("/:id/status", h.vendors.UpdateProfileStatus)

			vendorsGroup.POST("/locations", h.vendors.CreateLocation)
			vendorsGroup.GET("/locations/me", h.vendors.GetMyLocations)
			vendorsGroup.GET("/locations/:locationId", h.vendors.GetLocationByID)
			vendorsGroup.PATCH("/locations/:locationId", h.vendors.UpdateLocation)
			vendorsGroup.DELETE("/locations/:locationId", h.vendors.DeleteLocation)

			vendorsGroup.POST("/business-hours", h.vendors.SetBusinessHours)
			vendorsGroup.GET("/business-hours/me", h.vendors.GetBusinessHours)
		}

		// Customers routes
		customersGroup := protected.Group("/customers")
		{
			customersGroup.GET("/profile/me", h.customers.GetMyProfile)
			customersGroup.PATCH("/profile/me", h.customers.UpdateMyProfile)

			customersGroup.POST("/addresses", h.customers.CreateAddress)
			customersGroup.GET("/addresses/me", h.customers.GetMyAddresses)
			customersGroup.GET("/addresses/default", h.customers.GetDefaultAddress)
			customersGroup.GET("/addresses/:addressId", h.customers.GetAddressByID)
			customersGroup.PATCH("/addresses/:addressId", h.customers.UpdateAddress)
			customersGroup.DELETE("/addresses/:addressId", h.customers.DeleteAddress)
		}

		// Products routes (management operations)
		productsGroup := protected.Group("/products")
		{
			productsGroup.POST("", h.products.CreateProduct)
			productsGroup.GET("/active", h.products.GetMyActiveProducts)
			productsGroup.PATCH("/:id", h.products.UpdateProduct)
			productsGroup.DELETE("/:id", h.products.ArchiveProduct)
		}

		// Categories routes (management operations)
		categoriesGroup := protected.Group("/categories")
		{
			categoriesGroup.POST("", h.products.CreateCategory)
			categoriesGroup.PATCH("/:id", h.products.UpdateCategory)
		}

		// Inventory routes
		inventoryGroup := protected.Group("/inventory")
		{
			inventoryGroup.GET("/:productId", h.inventory.GetInventory)
			inventoryGroup.GET("/me", h.inventory.GetMyInventory)
			inventoryGroup.POST("/adjust", h.inventory.AdjustStock)
			inventoryGroup.POST("/reserve", h.inventory.ReserveStock)
			inventoryGroup.POST("/release", h.inventory.ReleaseStock)
			inventoryGroup.POST("/fulfil", h.inventory.FulfilStock)
			inventoryGroup.GET("/:productId/movements", h.inventory.GetMovements)
		}

		// Cart routes
		cartGroup := protected.Group("/carts")
		{
			cartGroup.GET("/me", h.carts.GetMyCart)
			cartGroup.DELETE("/me", h.carts.ClearCart)
			cartGroup.POST("/items", h.carts.AddItem)
			cartGroup.PATCH("/items/:productId", h.carts.UpdateItemQuantity)
			cartGroup.DELETE("/items/:productId", h.carts.RemoveItem)
			cartGroup.GET("/:id", h.carts.GetCart)
		}

		// Orders routes
		ordersGroup := protected.Group("/orders")
		{
			ordersGroup.POST("", h.orders.Checkout)
			ordersGroup.GET("", h.orders.ListOrders)
			ordersGroup.GET("/:id", h.orders.GetOrderByID)
			ordersGroup.POST("/:id/accept", h.orders.AcceptOrder)
			ordersGroup.POST("/:id/reject", h.orders.RejectOrder)
			ordersGroup.POST("/:id/prepare", h.orders.PrepareOrder)
			ordersGroup.POST("/:id/ready", h.orders.ReadyForPickup)
			ordersGroup.POST("/:id/out-for-delivery", h.orders.OutForDelivery)
			ordersGroup.POST("/:id/complete", h.orders.CompleteOrder)
			ordersGroup.POST("/:id/request-cancellation", h.orders.RequestCancellation)
			ordersGroup.POST("/:id/cancel", h.orders.RequestCancellation)
			ordersGroup.POST("/:id/decide-cancellation", h.orders.DecideCancellation)
		}

		// Payments routes
		paymentsGroup := protected.Group("/payments")
		{
			paymentsGroup.POST("", h.payments.RecordPayment)
			paymentsGroup.GET("/order/:orderId", h.payments.GetOrderPayments)
			paymentsGroup.POST("/:id/verify", h.payments.VerifyPayment)
			paymentsGroup.POST("/:id/refunds", h.payments.RecordRefund)
			paymentsGroup.GET("/:id/refunds", h.payments.GetRefunds)
			paymentsGroup.POST("/gateway/initiate", h.payments.InitiateGatewayPayment)
			paymentsGroup.POST("/gateway/verify", h.payments.VerifyGatewayPayment)
		}

		// Udhaar routes
		udhaarGroup := protected.Group("/udhaar")
		{
			udhaarGroup.POST("/accounts", h.udhaar.CreateAccount)
			udhaarGroup.GET("/accounts", h.udhaar.ListAccounts)
			udhaarGroup.GET("/accounts/:id", h.udhaar.GetAccount)
			udhaarGroup.PATCH("/accounts/:id/status", h.udhaar.UpdateAccountStatus)
			udhaarGroup.POST("/accounts/:id/entries", h.udhaar.AddEntry)
			udhaarGroup.GET("/accounts/:id/entries", h.udhaar.GetEntries)
		}

		// Notifications routes
		notifGroup := protected.Group("/notifications")
		{
			notifGroup.GET("", h.notifications.GetMyNotifications)
			notifGroup.POST("/read-all", h.notifications.MarkAllRead)
			notifGroup.POST("/:id/read", h.notifications.MarkRead)
			notifGroup.POST("/device-tokens", h.notifications.RegisterDeviceToken)
			notifGroup.DELETE("/device-tokens/:token", h.notifications.RevokeDeviceToken)
		}

		// Sync routes (offline operation intake)
		syncGroup := protected.Group("/sync")
		{
			syncGroup.POST("/operations", h.sync.SubmitOperation)
			syncGroup.GET("/operations", h.sync.GetMyOperations)
		}

		// Analytics routes
		analyticsGroup := protected.Group("/analytics")
		{
			analyticsGroup.GET("/vendor/summary", h.analytics.GetVendorSummary)
			analyticsGroup.GET("/vendor/detailed", h.analytics.GetDetailedVendorAnalytics)
			analyticsGroup.GET("/vendor/demand-forecast", h.analytics.GetDemandForecast)
			analyticsGroup.POST("/expenses", h.analytics.CreateExpense)
			analyticsGroup.GET("/expenses", h.analytics.GetExpenses)
			analyticsGroup.DELETE("/expenses/:id", h.analytics.DeleteExpense)
			analyticsGroup.GET("/weather", h.analytics.GetWeather)
			analyticsGroup.GET("/platform/overview", h.analytics.GetPlatformOverview)
		}

		// Reviews routes
		reviewsGroup := protected.Group("/reviews")
		{
			reviewsGroup.POST("", h.reviews.CreateReview)
			reviewsGroup.GET("/vendor/:vendorId", h.reviews.GetVendorReviews)
			reviewsGroup.GET("/admin", middleware.RequireRole("ADMIN", "SUPER_ADMIN"), h.reviews.GetAllReviews)
			reviewsGroup.PATCH("/:id/moderate", middleware.RequireRole("ADMIN", "SUPER_ADMIN"), h.reviews.ModerateReview)
		}

		// Loyalty routes
		loyaltyGroup := protected.Group("/loyalty")
		{
			loyaltyGroup.GET("/me", h.reviews.GetMyLoyalty)
		}

		// Audit routes (Admin only)
		auditGroup := protected.Group("/audit-logs")
		auditGroup.Use(middleware.RequireRole("ADMIN", "SUPER_ADMIN"))
		{
			auditGroup.GET("", h.audit.ListAuditLogs)
		}

		// Delivery routes
		deliveryGroup := protected.Group("/delivery")
		{
			deliveryGroup.POST("/orders/:orderId/assign", middleware.RequireRole("ADMIN", "SUPER_ADMIN"), h.delivery.AssignDelivery)
			deliveryGroup.GET("/admin/assignments", middleware.RequireRole("ADMIN", "SUPER_ADMIN"), h.delivery.ListAllAssignments)
			deliveryGroup.GET("/agent/assignments", middleware.RequireRole("DELIVERY_AGENT"), h.delivery.GetMyAssignments)
			deliveryGroup.PATCH("/orders/:orderId/status", middleware.RequireRole("DELIVERY_AGENT"), h.delivery.UpdateDeliveryStatus)
			deliveryGroup.POST("/orders/:orderId/location", middleware.RequireRole("DELIVERY_AGENT"), h.delivery.UpdateLocation)
			deliveryGroup.GET("/orders/:orderId/assignment", h.delivery.GetAssignment)
			deliveryGroup.GET("/orders/:orderId/location", h.delivery.GetLocation)
		}
	}

	return router
}

func (a *Application) Run() error {
	go func() {
		logger.Logger.Info().Str("addr", a.cfg.HTTPAddr).Msg("starting HTTP server")
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error().Err(err).Msg("HTTP server error")
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		logger.Logger.Error().Err(err).Msg("server forced to shutdown")
		return err
	}

	a.dbPool.Close()
	logger.Logger.Info().Msg("server exited")

	return nil
}
