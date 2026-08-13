module github.com/hackelia-micrantha/calathea-community

go 1.26.0

require modernc.org/sqlite v1.56.0

// modernc.org/sqlite documents modernc.org/libc as a fragile ABI dependency
// that should be pinned to the exact version used by the selected sqlite release.
require modernc.org/libc v1.74.4 // indirect
