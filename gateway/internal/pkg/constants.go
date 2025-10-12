package pkg

var (
	FILE_FOLDER = GetEnv("FILE_FOLDER", "./app/fileDump")
	PORT        = GetEnv("PORT", "4000")
	CORS_ORIGIN = GetEnv("CORS_ORIGIN", "http://localhost:3000")
)
