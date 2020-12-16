package error

type Error struct {
	code int
	message string
}

func New(info map[string]interface{}) Error{
	var error Error
	if info["message"] != nil {
		error.message = info["message"].(string)
	}

	return error
}