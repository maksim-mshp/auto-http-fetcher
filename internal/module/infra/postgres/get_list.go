package postgres

import (
	coreHttp "auto-http-fetcher/internal/core/http"
	domainModule "auto-http-fetcher/internal/module/domain"
	domainWebhook "auto-http-fetcher/internal/webhook/domain"

	"context"
)

func (r *PGModuleRepo) GetModuleList(ctx context.Context, userID int) ([]*domainModule.Module, error) {
	modulesQuery := `SELECT id, owner_id, name, description
                     FROM modules WHERE owner_id = $1 ORDER BY id DESC`

	rows, err := r.pool.Query(ctx, modulesQuery, userID)
	if err != nil {
		return nil, coreHttp.ErrInternal
	}
	defer rows.Close()

	var modules []*domainModule.Module
	var moduleIDs []int

	for rows.Next() {
		module := &domainModule.Module{}
		err = rows.Scan(
			&module.ID, &module.OwnerId, &module.Name, &module.Description)
		if err != nil {
			return nil, coreHttp.ErrInternal
		}
		modules = append(modules, module)
		moduleIDs = append(moduleIDs, module.ID)
	}

	if err = rows.Err(); err != nil {
		return nil, coreHttp.ErrInternal
	}

	if len(moduleIDs) == 0 {
		return modules, nil
	}

	webhooks, err := r.getWebhooksByModuleIDs(ctx, moduleIDs)
	if err != nil {
		return nil, coreHttp.ErrInternal
	}

	webhooksMap := make(map[int][]*domainWebhook.Webhook)
	for _, webhook := range webhooks {
		webhooksMap[webhook.ModuleID] = append(webhooksMap[webhook.ModuleID], webhook)
	}

	for _, module := range modules {
		if wh, ok := webhooksMap[module.ID]; ok {
			module.Webhooks = wh
		} else {
			module.Webhooks = []*domainWebhook.Webhook{}
		}
	}

	return modules, nil
}
