package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"gocms/internal/user"
	"os"
	"strconv"
	"strings"
)

func ReadString(prompt string) (string, error) {
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

func ReadBytes(prompt string) ([]byte, error) {
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	return []byte(input), nil
}

func CreateUser(ctx context.Context, app *Container) error {
	svc := app.Users.Service

	var username string
	var passwd []byte
	var permissions uint32

	for username == "" {
		input, err := ReadString("Enter username")
		if err != nil {
			logger.Error("User create", "Failed to read input string", err)
			return err
		}

		username, err = domain.NormalizeUsername(input)
		if err != nil {
			if !errors.Is(err, domain.ErrInval) {
				logger.Error("User create", "Failed to verify username", err)
				return err
			}
			fmt.Printf("%v", err)
		}
	}

	for permissions == 0 {
		input, err := ReadString("Enter user's permissions")
		if err != nil {
			logger.Error("User create", "Failed to read input string", err)
			return err
		}

		if input == "all" {
			permissions = uint32(4294967295)
		} else {
			val, err := strconv.ParseUint(input, 10, 32)
			if err != nil || val == 0 {
				fmt.Printf("Permissions value must be positive nonnull integer\n")
			} else {
				permissions = uint32(val)
			}
		}
	}

	for len(passwd) == 0 {
		input, err := ReadBytes("Enter password")
		if err != nil {
			logger.Error("User create", "Failed to read input string", err)
			return err
		}

		confirm, err := ReadBytes("Re-enter password")
		if err != nil {
			logger.Error("User create", "Failed to read input string", err)
			return err
		}

		if bytes.Equal(input, confirm) {
			clear(confirm)
			passwd = input
		} else {
			fmt.Printf("Passwords do not match\n")
			clear(input)
			clear(confirm)
		}
	}

	user := user.Record{
		Username:    username,
		Permissions: permissions,
		Passwd:      passwd,
	}

	injected := context.WithValue(ctx, auth.UserPermissionKey, domain.OnEntityC("users"))

	if err := svc.Create(injected, &user); err != nil {
		logger.Error("User create", "Failed to register new user", err)
		return err
	}
	clear(passwd)
	fmt.Println("User was created successfully")
	return nil
}
