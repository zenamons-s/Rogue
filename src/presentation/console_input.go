// Пакет presentation содержит консольный интерфейс игры (raw‑mode и line‑mode).
package presentation

import (
	"syscall"
	"unsafe"
)

// enterRawMode переводит терминал в raw‑mode (отключает буферизацию и эхо).
func (a *ConsoleApp) enterRawMode() error {
	fd := int(a.stdin.Fd())
	oldState, err := getTermios(fd)
	if err != nil {
		return err
	}
	raw := *oldState
	raw.Lflag &^= syscall.ICANON | syscall.ECHO
	raw.Iflag &^= syscall.ICRNL | syscall.INLCR
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if err := setTermios(fd, &raw); err != nil {
		return err
	}
	a.rawState = oldState
	return nil
}

// restoreTerminal восстанавливает исходные настройки терминала.
func (a *ConsoleApp) restoreTerminal() {
	a.restoreOnce.Do(func() {
		if a.rawState != nil {
			_ = setTermios(int(a.stdin.Fd()), a.rawState)
			a.rawState = nil
		}
	})
}

// readKey читает один символ из stdin (raw‑mode).
func (a *ConsoleApp) readKey() (rune, error) {
	var b [1]byte
	for {
		_, err := a.stdin.Read(b[:])
		if err != nil {
			return 0, err
		}
		ch := b[0]
		if ch == 0x1b || ch == '\r' || ch == '\n' {
			return 0, nil
		}
		if ch >= 'A' && ch <= 'Z' {
			ch += 'a' - 'A'
		}
		return rune(ch), nil
	}
}

// readControlKey читает один управляющий символ (raw‑mode), преобразуя заглавные буквы в строчные.
func (a *ConsoleApp) readControlKey() (rune, error) {
	var b [1]byte
	_, err := a.stdin.Read(b[:])
	if err != nil {
		return 0, err
	}
	ch := b[0]
	if ch >= 'A' && ch <= 'Z' {
		ch += 'a' - 'A'
	}
	if ch == 0x1b {
		return 0x1b, nil
	}
	if ch == '\r' || ch == '\n' {
		return '\n', nil
	}
	return rune(ch), nil
}

// getTermios получает текущие настройки терминала через системный вызов ioctl.
func getTermios(fd int) (*syscall.Termios, error) {
	state := &syscall.Termios{}
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(state)), 0, 0, 0)
	if errno != 0 {
		return nil, errno
	}
	return state, nil
}

// setTermios применяет новые настройки терминала через системный вызов ioctl.
func setTermios(fd int, state *syscall.Termios) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(state)), 0, 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}