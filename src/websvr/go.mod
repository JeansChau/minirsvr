module main

go 1.20

require voxwebsvr.com/client v0.0.0-00010101000000-000000000000

require voxwebsvr.com/webfs v0.0.0-00010101000000-000000000000 // indirect

replace voxwebsvr.com/webfs => ./webfs

replace voxwebsvr.com/client => ./client
