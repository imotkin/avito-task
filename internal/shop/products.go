package shop

type ProductType int

const (
	ProductTShirt ProductType = iota + 1
	ProductCup
	ProductBook
	ProductPen
	ProductPowerbank
	ProductHoody
	ProductUmbrella
	ProductSocks
	ProductWallet
	ProductPinkHoody
)

var names = [...]string{
	"t-shirt",
	"cup",
	"book",
	"pen",
	"powerbank",
	"hoody",
	"umbrella",
	"socks",
	"wallet",
	"pink-hoody",
}

var products map[string]struct{}

func init() {
	products = make(map[string]struct{}, len(names))
	for _, name := range names {
		products[name] = struct{}{}
	}
}

func IsProduct(product string) bool {
	_, ok := products[product]
	return ok
}
