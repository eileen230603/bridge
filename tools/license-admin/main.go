package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/symphonylicensemanager/go/license"
)

func issue(key ed25519.PrivateKey, id, expiration string, now time.Time) (string, error) {
	id = strings.ToLower(strings.TrimSpace(id))
	decoded, err := hex.DecodeString(id)
	if err != nil || len(decoded) != 32 {
		return "", errors.New("Machine ID debe tener 64 caracteres hexadecimales")
	}
	l := license.License{Product: "symphony-ap1", MachineID: id, IssuedAt: now.UTC()}
	if expiration == "" || strings.EqualFold(expiration, "permanente") {
		l.Indefinite = true
	} else {
		day, err := time.Parse("2006-01-02", expiration)
		if err != nil {
			return "", errors.New("fecha invalida: use AAAA-MM-DD o permanente")
		}
		l.ExpiresAt = day.Add(24*time.Hour - time.Nanosecond)
		if !l.ExpiresAt.After(now) {
			return "", errors.New("el vencimiento debe estar en el futuro")
		}
	}
	payload, err := json.Marshal(l)
	if err != nil {
		return "", err
	}
	signature := ed25519.Sign(key, payload)
	return base64.RawURLEncoding.EncodeToString(append(payload, signature...)), nil
}

func initialize(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "issuer-private.key")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = f.WriteString(base64.RawURLEncoding.EncodeToString(private))
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.WriteFile(filepath.Join(dir, "issuer-public.key"), []byte(base64.RawURLEncoding.EncodeToString(public)), 0644)
}

func loadKey(dir string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(filepath.Join(dir, "issuer-private.key"))
	if err != nil {
		return nil, err
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(string(data)))
	if err != nil || len(raw) != ed25519.PrivateKeySize {
		return nil, errors.New("clave privada invalida")
	}
	key := ed25519.NewKeyFromSeed(raw[:ed25519.SeedSize])
	if !key.Equal(ed25519.PrivateKey(raw)) {
		return nil, errors.New("clave privada inconsistente")
	}
	return key, nil
}

func run() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dir := flag.String("keys", filepath.Dir(exe), "Carpeta privada del administrador")
	initKeys := flag.Bool("init", false, "Crear un par de claves nuevo (solo configuracion inicial)")
	id := flag.String("machine", "", "Machine ID")
	expires := flag.String("expires", "permanente", "AAAA-MM-DD (fin del dia UTC) o permanente")
	flag.Parse()
	if *initKeys {
		return initialize(*dir)
	}
	if *id == "" {
		return runGUI(*dir)
	}
	key, err := loadKey(*dir)
	if err != nil {
		return fmt.Errorf("no se pudo cargar la clave del administrador: %w", err)
	}
	token, err := issue(key, *id, *expires, time.Now())
	if err != nil {
		return err
	}
	admin, err := newAdmin(*dir)
	if err != nil {
		return err
	}
	defer admin.db.Close()
	details, err := decodeIssued(token, key.Public().(ed25519.PublicKey))
	if err != nil {
		return err
	}
	if err := admin.save("Emitida por consola", token, details); err != nil {
		return err
	}
	output := filepath.Join(*dir, "licencia-"+strings.ToLower(strings.TrimSpace(*id))+"-"+time.Now().UTC().Format("20060102-150405.000000000")+".txt")
	if err := os.WriteFile(output, []byte(token+"\n"), 0600); err != nil {
		return err
	}
	fmt.Println("\nToken de licencia:\n" + token)
	fmt.Println("\nGuardado en: " + output)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)

		os.Exit(1)
	}
}
