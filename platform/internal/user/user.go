package user

// User 用戶資料結構
type User struct {
	ID       string
	Username string
	Email    string
	Balance  float64
	Role     string
}

// Manager 用戶管理器（Demo 版本）
type Manager struct {
	users map[string]*User
}

// NewManager 創建新的用戶管理器
func NewManager() *Manager {
	// Demo 階段：預先創建一些測試用戶
	users := map[string]*User{
		"demo-user-001": {
			ID:       "demo-user-001",
			Username: "demo_user",
			Email:    "demo@golden-buy.com",
			Balance:  10000.00,
			Role:     "demo",
		},
		"test-user-001": {
			ID:       "test-user-001",
			Username: "test_trader",
			Email:    "test@golden-buy.com",
			Balance:  50000.00,
			Role:     "premium",
		},
	}

	return &Manager{
		users: users,
	}
}

// GetUser 獲取用戶資訊
func (m *Manager) GetUser(id string) (*User, bool) {
	user, exists := m.users[id]
	return user, exists
}

// GetDefaultUser 獲取預設用戶（Demo 版本）
func (m *Manager) GetDefaultUser() *User {
	return m.users["demo-user-001"]
}

// UpdateBalance 更新用戶餘額
func (m *Manager) UpdateBalance(id string, amount float64) error {
	user, exists := m.users[id]
	if !exists {
		return ErrUserNotFound
	}

	user.Balance += amount
	return nil
}

// SetBalance 設置用戶餘額
func (m *Manager) SetBalance(id string, balance float64) error {
	user, exists := m.users[id]
	if !exists {
		return ErrUserNotFound
	}

	user.Balance = balance
	return nil
}

// GetBalance 獲取用戶餘額
func (m *Manager) GetBalance(id string) (float64, error) {
	user, exists := m.users[id]
	if !exists {
		return 0, ErrUserNotFound
	}

	return user.Balance, nil
}

// ListUsers 列出所有用戶（Demo 版本）
func (m *Manager) ListUsers() []*User {
	users := make([]*User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users
}
