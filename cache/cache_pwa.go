package cache

import (
	"context"
	"encoding/base64"
	"r3/db"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	pwa_mx       sync.RWMutex
	pwaIconIdMap = make(map[uuid.UUID]string)
	pwaDomainMap = make(map[string]uuid.UUID) // key = sub domain name, value = module ID (for direct app acess)
)

func GetPwaIcon(id uuid.UUID) (string, error) {
	pwa_mx.RLock()
	file, exists := pwaIconIdMap[id]
	pwa_mx.RUnlock()

	if exists {
		return file, nil
	}

	var f []byte
	if err := db.Pool.QueryRow(context.Background(), `
		SELECT file
		FROM app.icon
		WHERE id = $1
	`, id).Scan(&f); err != nil {
		return file, err
	}
	file = base64.StdEncoding.EncodeToString(f)

	pwa_mx.Lock()
	pwaIconIdMap[id] = file
	pwa_mx.Unlock()

	return file, nil
}

func GetPwaDomainMap() map[string]uuid.UUID {
	pwa_mx.RLock()
	defer pwa_mx.RUnlock()

	return pwaDomainMap
}

func GetPwaModuleId(subdomain string) uuid.UUID {
	pwa_mx.RLock()
	defer pwa_mx.RUnlock()

	id, exists := pwaDomainMap[subdomain]
	if !exists {
		return uuid.Nil
	}
	return id
}

func LoadPwaDomainMap_tx(ctx context.Context, tx pgx.Tx) error {
	pwa_mx.Lock()
	defer pwa_mx.Unlock()

	pwaDomainMap = make(map[string]uuid.UUID)

	rows, err := tx.Query(ctx, `
		SELECT module_id, domain
		FROM instance.pwa_domain
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var modId uuid.UUID
		var domain string

		if err := rows.Scan(&modId, &domain); err != nil {
			return err
		}
		pwaDomainMap[domain] = modId
	}
	return nil
}
