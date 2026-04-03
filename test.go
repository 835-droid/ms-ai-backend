package main

import (
	"fmt"
	"time"

	u "github.com/835-droid/ms-ai-backend/internal/core/user"
)

func main() {
	ud := u.UserDetails{CreatedAt: time.Now()}
	fmt.Println(ud.CreatedAt)
}
