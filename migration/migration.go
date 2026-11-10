package migration

import "embed"

//go:embed postgres/*sql
var Postgres embed.FS
