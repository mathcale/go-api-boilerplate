package repositories

var (
	queryUserExists     string
	queryGetUserByEmail string
	queryInsertUser     string
)

func init() {
	queryUserExists = `SELECT count(*) > 0 AS exists FROM users WHERE email = $1 AND active IS TRUE`
	queryGetUserByEmail = `SELECT * FROM users WHERE email = $1 AND active IS TRUE`
	queryInsertUser = `INSERT INTO users (name, email, password, active) VALUES ($1, $2, $3, $4)`
}
