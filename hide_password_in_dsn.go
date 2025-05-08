package main

import (
	"fmt"
	"strings"
)

func hidePasswordInDSN(dsn string) string {
	parts := strings.SplitN(dsn, "://", 2)
	if len(parts) != 2 {
		return dsn
	}

	authAndRest := strings.SplitN(parts[1], "@", 2)
	if len(authAndRest) != 2 {
		return dsn
	}

	userAndPass := strings.SplitN(authAndRest[0], ":", 2)
	if len(userAndPass) != 2 {
		return dsn
	}

	return fmt.Sprintf("%s://%s:*****@%s", parts[0], userAndPass[0], authAndRest[1])
}
