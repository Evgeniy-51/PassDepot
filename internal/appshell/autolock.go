package appshell

import (
	"context"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	defaultAutoLockMinutes = 15
	autoLockTick           = 15 * time.Second
)

func (a *App) stopAutoLockLocked() {
	if a.autoLockCancel != nil {
		a.autoLockCancel()
		a.autoLockCancel = nil
	}
}

// startAutoLockLocked запускает фоновую проверку простоя (вызывать уже под a.mu).
func (a *App) startAutoLockLocked() {
	a.stopAutoLockLocked()
	if a.profileID == "" || a.autoLockMinutes <= 0 {
		return
	}
	a.lastActivity = time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	a.autoLockCancel = cancel
	go a.autoLockLoop(ctx)
}

func (a *App) autoLockLoop(ctx context.Context) {
	tick := time.NewTicker(autoLockTick)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			a.mu.Lock()
			if a.profileID == "" {
				a.mu.Unlock()
				return
			}
			if a.autoLockMinutes <= 0 {
				a.mu.Unlock()
				return
			}
			if time.Since(a.lastActivity) <= time.Duration(a.autoLockMinutes)*time.Minute {
				a.mu.Unlock()
				continue
			}
			a.logoutLocked()
			a.mu.Unlock()
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "session:locked", "")
			}
			return
		}
	}
}

// touchActivityLocked обновляет время активности (уже под a.mu).
func (a *App) touchActivityLocked() {
	if a.profileID != "" {
		a.lastActivity = time.Now()
	}
}

// GetAutoLockMinutes возвращает таймаут авто-блокировки в минутах (0 = выкл).
func (a *App) GetAutoLockMinutes() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.autoLockMinutes
}

// SetAutoLockMinutes задаёт таймаут (0…240; 0 = отключить). При активной сессии перезапускает таймер.
func (a *App) SetAutoLockMinutes(minutes int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if minutes < 0 {
		minutes = 0
	}
	if minutes > 240 {
		minutes = 240
	}
	a.autoLockMinutes = minutes
	if a.profileID != "" {
		a.startAutoLockLocked()
	}
}
