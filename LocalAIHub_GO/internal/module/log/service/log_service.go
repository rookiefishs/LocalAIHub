package service

import (
	"context"
	"io"

	gatewayrepo "localaihub/localaihub_go/internal/module/gateway/repository"
	logrepo "localaihub/localaihub_go/internal/module/log/repository"
)

type LogService struct {
	requestRepo *gatewayrepo.GatewayRepository
	auditRepo   *logrepo.LogRepository
}

func NewLogService(requestRepo *gatewayrepo.GatewayRepository, auditRepo *logrepo.LogRepository) *LogService {
	return &LogService{requestRepo: requestRepo, auditRepo: auditRepo}
}

func (s *LogService) ListRequestLogs(ctx context.Context, filters gatewayrepo.RequestLogFilters) ([]gatewayrepo.RequestLogRecord, int, error) {
	return s.requestRepo.ListRequestLogs(ctx, filters)
}

func (s *LogService) GetRequestLog(ctx context.Context, id int64) (*gatewayrepo.RequestLogRecord, error) {
	return s.requestRepo.GetRequestLogByID(ctx, id)
}

func (s *LogService) ListAuditLogs(ctx context.Context, filters logrepo.AuditLogFilters) ([]logrepo.AuditLogRecord, int, error) {
	return s.auditRepo.ListAuditLogs(ctx, filters)
}

func (s *LogService) GetAuditLog(ctx context.Context, id int64) (*logrepo.AuditLogRecord, error) {
	return s.auditRepo.GetAuditLogByID(ctx, id)
}

func (s *LogService) ExportAuditLogsCSV(ctx context.Context, filters logrepo.AuditLogFilters, writer io.Writer) error {
	return s.auditRepo.ExportAuditLogsCSV(ctx, filters, writer)
}
