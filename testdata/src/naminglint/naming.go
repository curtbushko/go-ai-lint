package naminglint

// ===== AIL050: getter-with-get-prefix =====

// User is a sample struct for testing getters.
type User struct {
	name  string
	email string
	age   int
}

// GetName is a bad getter - should be Name().
func (u *User) GetName() string { // want "AIL050: getter should not have Get prefix"
	return u.name
}

// GetEmail is a bad getter - should be Email().
func (u *User) GetEmail() string { // want "AIL050: getter should not have Get prefix"
	return u.email
}

// GetAge is a bad getter - should be Age().
func (u *User) GetAge() int { // want "AIL050: getter should not have Get prefix"
	return u.age
}

// Name is a good getter - no Get prefix.
func (u *User) Name() string {
	return u.name
}

// Email is a good getter - no Get prefix.
func (u *User) Email() string {
	return u.email
}

// GetUserByID is OK - not a simple getter, takes parameters.
func GetUserByID(id string) *User {
	return &User{}
}

// GetOrCreate is OK - not a simple getter, has side effects implied.
func (u *User) GetOrCreate() *User {
	return u
}

// GetContext is OK - returns context (common pattern).
func (u *User) GetContext() string {
	return ""
}

// ===== AIL051: redundant-package-prefix =====

// NaminglintService has redundant package prefix.
type NaminglintService struct{} // want "AIL051: type name repeats package name"

// NaminglintConfig has redundant package prefix.
type NaminglintConfig struct{} // want "AIL051: type name repeats package name"

// NaminglintError has redundant package prefix.
type NaminglintError struct{} // want "AIL051: type name repeats package name"

// Service is good - no redundant prefix.
type Service struct{}

// Config is good - no redundant prefix.
type Config struct{}

// UserService is OK - different prefix than package name.
type UserService struct{}

// NamingHelper is OK - "Naming" is not the same as "naminglint".
type NamingHelper struct{}

// naminglintInternal is unexported, so we don't flag it.
type naminglintInternal struct{}
