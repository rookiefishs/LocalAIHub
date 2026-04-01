package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"localaihub/localaihub_go/internal/app/router"
	"localaihub/localaihub_go/internal/config"
	"localaihub/localaihub_go/internal/database"
	auditrepo "localaihub/localaihub_go/internal/module/audit/repository"
	auditservice "localaihub/localaihub_go/internal/module/audit/service"
	authhandler "localaihub/localaihub_go/internal/module/auth/handler"
	authrepo "localaihub/localaihub_go/internal/module/auth/repository"
	authservice "localaihub/localaihub_go/internal/module/auth/service"
	clientkeyhandler "localaihub/localaihub_go/internal/module/clientkey/handler"
	clientkeyrepo "localaihub/localaihub_go/internal/module/clientkey/repository"
	clientkeyservice "localaihub/localaihub_go/internal/module/clientkey/service"
	configexporthandler "localaihub/localaihub_go/internal/module/configexport/handler"
	configexportservice "localaihub/localaihub_go/internal/module/configexport/service"
	gatewayhandler "localaihub/localaihub_go/internal/module/gateway/handler"
	gatewayrepo "localaihub/localaihub_go/internal/module/gateway/repository"
	gatewayservice "localaihub/localaihub_go/internal/module/gateway/service"
	loghandler "localaihub/localaihub_go/internal/module/log/handler"
	logrepo "localaihub/localaihub_go/internal/module/log/repository"
	logservice "localaihub/localaihub_go/internal/module/log/service"
	modelhandler "localaihub/localaihub_go/internal/module/model/handler"
	modelrepo "localaihub/localaihub_go/internal/module/model/repository"
	modelservice "localaihub/localaihub_go/internal/module/model/service"
	providerhandler "localaihub/localaihub_go/internal/module/provider/handler"
	providerkeyrepo "localaihub/localaihub_go/internal/module/provider/repository"
	providerrepo "localaihub/localaihub_go/internal/module/provider/repository"
	providerservice "localaihub/localaihub_go/internal/module/provider/service"
	routehandler "localaihub/localaihub_go/internal/module/route/handler"
	routerepo "localaihub/localaihub_go/internal/module/route/repository"
	routeservice "localaihub/localaihub_go/internal/module/route/service"
	toolshandler "localaihub/localaihub_go/internal/module/tools/handler"
)

type App struct {
	Config config.Config
	DB     *sql.DB
	Server *http.Server
}

func New() (*App, error) {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	db, err := database.NewMySQL(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("init mysql: %w", err)
	}

	passwordHash := cfg.Security.AdminPasswordHash
	if cfg.Security.AdminPasswordPlain != "" {
		hashed, hashErr := bcrypt.GenerateFromPassword([]byte(cfg.Security.AdminPasswordPlain), bcrypt.DefaultCost)
		if hashErr != nil {
			return nil, fmt.Errorf("generate admin password hash from plain config: %w", hashErr)
		}
		passwordHash = string(hashed)
	}
	if passwordHash == "" {
		hashed, hashErr := bcrypt.GenerateFromPassword([]byte("yu3209605851"), bcrypt.DefaultCost)
		if hashErr != nil {
			return nil, fmt.Errorf("generate default admin password hash: %w", hashErr)
		}
		passwordHash = string(hashed)
	}

	adminRepo := authrepo.NewAdminRepository(db)
	if err := adminRepo.EnsureDefaultAdmin(ctx, "admin", passwordHash); err != nil {
		return nil, fmt.Errorf("ensure default admin: %w", err)
	}
	auditRepo := auditrepo.NewAuditRepository(db)
	auditSvc := auditservice.NewAuditService(auditRepo)
	authSvc := authservice.NewAuthService(adminRepo, cfg.Security.AdminSessionSecret)

	providerRepo := providerrepo.NewProviderRepository(db)
	providerKeyRepo := providerkeyrepo.NewProviderKeyRepository(db)
	providerKeySvc := providerservice.NewProviderKeyService(providerKeyRepo, cfg.Security.EncryptionKey, auditSvc)
	providerSvc := providerservice.NewProviderService(providerRepo, providerKeySvc, auditSvc)

	modelRepo := modelrepo.NewModelRepository(db)
	modelSvc := modelservice.NewModelService(modelRepo, auditSvc, providerSvc)

	gatewayRepo := gatewayrepo.NewGatewayRepository(db)
	clientKeyRepo := clientkeyrepo.NewClientKeyRepository(db)
	clientKeySvc := clientkeyservice.NewClientKeyService(clientKeyRepo, gatewayRepo, providerKeySvc, auditSvc)

	routeRepo := routerepo.NewRouteRepository(db)
	routeSvc := routeservice.NewRouteService(routeRepo, auditSvc)
	gatewaySvc := gatewayservice.NewGatewayService(gatewayRepo, providerKeySvc, routeSvc)
	logRepo := logrepo.NewLogRepository(db)
	logSvc := logservice.NewLogService(gatewayRepo, logRepo)

	configExportSvc := configexportservice.NewExportService(db)

	handlers := router.Handlers{
		Auth:        authhandler.NewAdminAuthHandler(authSvc),
		System:      routehandler.NewDashboardHandler(routeSvc, gatewayRepo, providerRepo),
		Proxy:       gatewayhandler.NewProxyHandler(gatewaySvc),
		Providers:   providerhandler.NewProviderHandler(providerSvc),
		ProviderKey: providerhandler.NewProviderKeyHandler(providerKeySvc),
		Models:      modelhandler.NewModelHandler(modelSvc),
		ClientKey:   clientkeyhandler.NewClientKeyHandler(clientKeySvc),
		Route:       routehandler.NewRouteHandler(routeSvc),
		Logs:        loghandler.NewLogHandler(logSvc),
		Tools:       toolshandler.NewToolsHandler(gatewaySvc),
		Config:      configexporthandler.NewConfigExportHandler(configExportSvc),
	}

	mux := router.New(handlers, authSvc, cfg.CORS.AllowedOrigins)
	server := &http.Server{
		Addr:              cfg.Server.Address(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		Config: cfg,
		DB:     db,
		Server: server,
	}, nil
}
