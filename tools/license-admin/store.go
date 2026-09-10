package main

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	wr "github.com/wailsapp/wails/v2/pkg/runtime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/symphonylicensemanager/go/license"
	_ "modernc.org/sqlite"
)

type LicenseRow struct {
	ID               int64  `json:"id"`
	Client           string `json:"client"`
	MachineID        string `json:"machineId"`
	Token            string `json:"token"`
	IssuedAt         string `json:"issuedAt"`
	ExpiresAt        string `json:"expiresAt"`
	Permanent        bool   `json:"permanent"`
	Status           string `json:"status"`
	RemainingSeconds int64  `json:"remainingSeconds"`
}
type IssueRequest struct {
	Client    string `json:"client"`
	MachineID string `json:"machineId"`
	Expires   string `json:"expires"`
	Permanent bool   `json:"permanent"`
}
type Admin struct {
	ctx        context.Context
	db         *sql.DB
	key        ed25519.PrivateKey
	startupErr error
}

func openStore(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`PRAGMA busy_timeout=5000;
 CREATE TABLE IF NOT EXISTS licenses (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 client TEXT NOT NULL DEFAULT '', machine_id TEXT NOT NULL,
 token TEXT NOT NULL UNIQUE, issued_at TEXT NOT NULL,
 expires_at TEXT NOT NULL DEFAULT '', permanent INTEGER NOT NULL CHECK(permanent IN (0,1))
 ); CREATE INDEX IF NOT EXISTS licenses_machine ON licenses(machine_id);`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
func decodeIssued(token string, public ed25519.PublicKey) (license.License, error) {
	var l license.License
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil || len(raw) < ed25519.SignatureSize {
		return l, errors.New("token inválido")
	}
	payload, sig := raw[:len(raw)-ed25519.SignatureSize], raw[len(raw)-ed25519.SignatureSize:]
	if !ed25519.Verify(public, payload, sig) {
		return l, errors.New("el token pertenece a otro emisor o está alterado")
	}
	if err := json.Unmarshal(payload, &l); err != nil {
		return l, err
	}
	if l.Product != "symphony-ap1" || l.MachineID == "" || l.IssuedAt.IsZero() || (!l.Indefinite && l.ExpiresAt.IsZero()) {
		return l, errors.New("datos de licencia incompletos")
	}
	return l, nil
}
func (a *Admin) save(client, token string, l license.License) error {
	expires := ""
	if !l.Indefinite {
		expires = l.ExpiresAt.UTC().Format(time.RFC3339Nano)
	}
	_, err := a.db.Exec(`INSERT INTO licenses(client,machine_id,token,issued_at,expires_at,permanent) VALUES(?,?,?,?,?,?) ON CONFLICT(token) DO NOTHING`, strings.TrimSpace(client), l.MachineID, token, l.IssuedAt.UTC().Format(time.RFC3339Nano), expires, l.Indefinite)
	return err
}
func newAdmin(dir string) (*Admin, error) {
	key, err := loadKey(dir)
	if err != nil {
		return nil, fmt.Errorf("No se encontró una clave de emisión válida junto al programa: %w", err)
	}
	db, err := openStore(filepath.Join(dir, "licenses.sqlite"))
	if err != nil {
		return nil, err
	}
	a := &Admin{db: db, key: key}
	files, err := filepath.Glob(filepath.Join(dir, "licencia-*.txt"))
	if err != nil {
		db.Close()
		return nil, err
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			db.Close()
			return nil, err
		}
		token := strings.TrimSpace(string(data))
		l, err := decodeIssued(token, key.Public().(ed25519.PublicKey))
		if err != nil {
			continue
		}
		if err := a.save("Licencia anterior", token, l); err != nil {
			db.Close()
			return nil, err
		}
	}
	return a, nil
}
func statusAt(row *LicenseRow, now time.Time) error {
	if row.Permanent {
		row.Status = "permanent"
		return nil
	}
	expiry, err := time.Parse(time.RFC3339Nano, row.ExpiresAt)
	if err != nil {
		return err
	}
	row.RemainingSeconds = int64(expiry.Sub(now).Seconds())
	if !expiry.After(now) {
		row.Status = "expired"
	} else if expiry.Sub(now) <= 30*24*time.Hour {
		row.Status = "soon"
	} else {
		row.Status = "active"
	}
	return nil
}
func (a *Admin) ListLicenses() ([]LicenseRow, error) {
	if a.startupErr != nil {
		return nil, a.startupErr
	}
	rows, err := a.db.Query(`SELECT id,client,machine_id,token,issued_at,expires_at,permanent FROM licenses ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []LicenseRow{}
	for rows.Next() {
		var l LicenseRow
		if err := rows.Scan(&l.ID, &l.Client, &l.MachineID, &l.Token, &l.IssuedAt, &l.ExpiresAt, &l.Permanent); err != nil {
			return nil, err
		}
		if err := statusAt(&l, time.Now()); err != nil {
			return nil, err
		}
		result = append(result, l)
	}
	return result, rows.Err()
}
func (a *Admin) CreateLicense(req IssueRequest) (LicenseRow, error) {
	if a.startupErr != nil {
		return LicenseRow{}, a.startupErr
	}
	req.Client = strings.TrimSpace(req.Client)
	if req.Client == "" || len(req.Client) > 160 {
		return LicenseRow{}, errors.New("Escribe un cliente o institución de hasta 160 caracteres")
	}
	expiration := req.Expires
	if req.Permanent {
		expiration = "permanente"
	} else if strings.TrimSpace(expiration) == "" {
		return LicenseRow{}, errors.New("Selecciona la fecha de vencimiento")
	}
	token, err := issue(a.key, req.MachineID, expiration, time.Now())
	if err != nil {
		return LicenseRow{}, err
	}
	l, err := decodeIssued(token, a.key.Public().(ed25519.PublicKey))
	if err != nil {
		return LicenseRow{}, err
	}
	if err := a.save(req.Client, token, l); err != nil {
		return LicenseRow{}, fmt.Errorf("No se pudo guardar la licencia: %w", err)
	}
	row := LicenseRow{Client: req.Client, MachineID: l.MachineID, Token: token, IssuedAt: l.IssuedAt.Format(time.RFC3339Nano), Permanent: l.Indefinite}
	if !l.Indefinite {
		row.ExpiresAt = l.ExpiresAt.Format(time.RFC3339Nano)
	}
	if err := a.db.QueryRow("SELECT id FROM licenses WHERE token=?", token).Scan(&row.ID); err != nil {
		return LicenseRow{}, err
	}
	return row, statusAt(&row, time.Now())
}

func (a *Admin) ExportLicense(id int64) (bool, error) {
	if a.startupErr != nil {
		return false, a.startupErr
	}
	var token, machine string
	if err := a.db.QueryRow("SELECT token,machine_id FROM licenses WHERE id=?", id).Scan(&token, &machine); err != nil {
		return false, err
	}
	path, err := wr.SaveFileDialog(a.ctx, wr.SaveDialogOptions{Title: "Guardar token de licencia", DefaultFilename: fmt.Sprintf("licencia-%s-%d.txt", machine, id), Filters: []wr.FileFilter{{DisplayName: "Token de licencia (*.txt)", Pattern: "*.txt"}}})
	if err != nil {
		return false, err
	}
	if path == "" {
		return false, nil
	}
	if err := os.WriteFile(path, []byte(token+"\n"), 0600); err != nil {
		return false, err
	}
	return true, nil
}
