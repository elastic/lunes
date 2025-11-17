module github.com/elastic/lunes/benchmarks

go 1.24.0

require (
	github.com/elastic/lunes v0.0.0-00010101000000-000000000000
	github.com/goodsign/monday v1.0.2
)

require github.com/magefile/mage v1.15.0 // indirect

replace github.com/elastic/lunes => ../
