package errorx

import "github.com/go-sql-driver/mysql"

func isCodeErr(err error, code Code) bool {
	if err == nil {
		return false
	}
	myErr, ok := err.(*myError)
	if !ok {
		return false
	}
	return myErr.code == code
}

func IsInternalErr(err error) bool {
	return isCodeErr(err, CodeInternal)
}

func IsNotFound(err error) bool {
	return isCodeErr(err, CodeNotFound)
}

func IsNoAuth(err error) bool {
	return isCodeErr(err, CodeNoAuth)
}

func IsDuplicateMySQLError(err error) bool {
	if err == nil {
		return false
	}

	mysqlErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}
	return mysqlErr.Number == 1062
}
