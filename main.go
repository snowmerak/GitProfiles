package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snowmerak/GitProfiles/key"
)

const envDir = "GITPROFILES_DIR"

// Init creates the base directory structure under baseDir
func Init(baseDir string) error {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		baseDir = filepath.Join(home, ".ssh", "git_profiles")
	}

	dirs := []string{
		filepath.Join(baseDir, "keys"),
		filepath.Join(baseDir, "meta"),
		filepath.Join(baseDir, "backups"),
		filepath.Join(baseDir, "gpg"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o700); err != nil {
			return err
		}
	}

	// ensure keys.json exists
	keysMeta := filepath.Join(baseDir, "meta", "keys.json")
	if _, err := os.Stat(keysMeta); os.IsNotExist(err) {
		if err := os.WriteFile(keysMeta, []byte("{}"), 0o600); err != nil {
			return err
		}
	}

	return nil
}

// Add generates a key using given algo and stores it under baseDir
func Add(baseDir, algo, name, email, host string) (privatePath, publicPath string, err error) {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", "", err
		}
		baseDir = filepath.Join(home, ".ssh", "git_profiles")
	}

	if algo == "" || name == "" || email == "" {
		return "", "", errors.New("algo, name and email are required")
	}

	gen, err := key.GetKeyGenerator(algo)
	if err != nil {
		return "", "", err
	}

	priv, pub, err := gen.Generate(name, email)
	if err != nil {
		return "", "", err
	}

	keysDir := filepath.Join(baseDir, "keys")
	if err := os.MkdirAll(keysDir, 0o700); err != nil {
		return "", "", err
	}

	baseName := fmt.Sprintf("%s_id_%s", name, strings.ReplaceAll(algo, "-", "_"))
	privatePath = filepath.Join(keysDir, baseName)
	publicPath = privatePath + ".pub"

	if err := os.WriteFile(privatePath, []byte(priv), 0o600); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(publicPath, []byte(pub), 0o644); err != nil {
		return "", "", err
	}

	// update meta
	metaPath := filepath.Join(baseDir, "meta", "keys.json")
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		return privatePath, publicPath, err
	}
	var meta map[string]map[string]string
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		meta = make(map[string]map[string]string)
	}

	meta[name] = map[string]string{
		"algo":    algo,
		"private": privatePath,
		"public":  publicPath,
		"email":   email,
		"host":    host,
	}

	out, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return privatePath, publicPath, err
	}
	if err := os.WriteFile(metaPath, out, 0o600); err != nil {
		return privatePath, publicPath, err
	}

	return privatePath, publicPath, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: gitprofiles <command> [flags]")
		fmt.Println("commands: init, add")
		os.Exit(2)
	}

	sub := os.Args[1]
	switch sub {
	case "init":
		initCmd := flag.NewFlagSet("init", flag.ExitOnError)
		base := initCmd.String("base", os.Getenv(envDir), "base directory for gitprofiles (overrides HOME)")
		initCmd.Parse(os.Args[2:])
		if err := Init(*base); err != nil {
			fmt.Fprintln(os.Stderr, "init error:", err)
			os.Exit(1)
		}
		fmt.Println("initialized")
	case "add":
		addCmd := flag.NewFlagSet("add", flag.ExitOnError)
		algo := addCmd.String("algo", "ed25519", "algorithm (ed25519, rsa2048, rsa4096, p256, p384, p521)")
		name := addCmd.String("name", "", "profile name")
		email := addCmd.String("email", "", "email/identity")
		host := addCmd.String("host", "", "host to use in ssh config (e.g. github.com)")
		base := addCmd.String("base", os.Getenv(envDir), "base directory for gitprofiles (overrides HOME)")
		addCmd.Parse(os.Args[2:])
		if *name == "" || *email == "" {
			addCmd.Usage()
			os.Exit(2)
		}
		priv, pub, err := Add(*base, *algo, *name, *email, *host)
		if err != nil {
			fmt.Fprintln(os.Stderr, "add error:", err)
			os.Exit(1)
		}
		fmt.Printf("private: %s\npublic: %s\n", priv, pub)
	case "ssh-config":
		if len(os.Args) < 3 {
			fmt.Println("usage: gitprofiles ssh-config <status|sync> [flags]")
			os.Exit(2)
		}
		sub := os.Args[2]
		switch sub {
		case "status":
			statusCmd := flag.NewFlagSet("ssh-config status", flag.ExitOnError)
			cfgPath := statusCmd.String("config", os.ExpandEnv("$HOME/.ssh/config"), "ssh config file path")
			base := statusCmd.String("base", os.Getenv(envDir), "base directory for gitprofiles (overrides HOME)")
			prune := statusCmd.Bool("prune", true, "show entries that would be removed if prune is enabled")
			statusCmd.Parse(os.Args[3:])
			adds, removes, err := PreviewSSHConfig(*base, *cfgPath, *prune)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ssh-config status error:", err)
				os.Exit(1)
			}
			if len(adds) == 0 && len(removes) == 0 {
				fmt.Println("ssh-config is up to date")
				return
			}
			if len(adds) > 0 {
				fmt.Println("Entries to add/update:")
				for _, e := range adds {
					fmt.Printf("  - alias: %s host: %s identity: %s\n", e.Alias, e.HostName, e.IdentityFile)
				}
			}
			if len(removes) > 0 {
				fmt.Println("Entries to remove:")
				for _, a := range removes {
					fmt.Printf("  - alias: %s\n", a)
				}
			}
		case "sync":
			syncCmd := flag.NewFlagSet("ssh-config sync", flag.ExitOnError)
			cfgPath := syncCmd.String("config", os.ExpandEnv("$HOME/.ssh/config"), "ssh config file path")
			base := syncCmd.String("base", os.Getenv(envDir), "base directory for gitprofiles (overrides HOME)")
			prune := syncCmd.Bool("prune", true, "remove stale managed entries not present in meta")
			syncCmd.Parse(os.Args[3:])
			if err := SyncSSHConfig(*base, *cfgPath, *prune); err != nil {
				fmt.Fprintln(os.Stderr, "ssh-config sync error:", err)
				os.Exit(1)
			}
			fmt.Println("ssh-config synced")
		default:
			fmt.Println("unknown ssh-config subcommand")
			os.Exit(2)
		}
	default:
		fmt.Println("unknown command")
		os.Exit(2)
	}
}
