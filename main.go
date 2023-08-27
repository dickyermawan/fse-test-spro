package main

import (
	"fmt"
	"net/http"
	"sync"
)

var (
	users = map[string]string{
		"user1": "password123",
	}
	balances = map[string]float64{
		"user1": 100,
	}
	products = map[string]float64{
		"apple":  1.0,
		"banana": 0.5,
	}
	cart  = make(map[string]map[string]int)
	mu    = &sync.Mutex{}
	minus = false
)

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add_to_cart/", addToCartHandler)
	http.HandleFunc("/checkout", checkoutHandler)
	http.ListenAndServe(":8080", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if storedPass, ok := users[username]; ok && storedPass == password {
			http.SetCookie(w, &http.Cookie{
				Name:  "session",
				Value: username,
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	fmt.Fprint(w, `<form action="/login" method="post">
Username: <input type="text" name="username" required><br>
Password: <input type="password" name="password" required><br>
<input type="submit" value="Login">
</form>`)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := cookie.Value
	userCart, ok := cart[username]
	if !ok {
		userCart = make(map[string]int)
		cart[username] = userCart
	}

	fmt.Fprintf(w, "Hello, %s! Available products:\n", username)
	for product, price := range products {
		fmt.Fprintf(w, "%s: $%.2f <a href=\"/add_to_cart/%s\">Add to cart</a>\n", product, price, product)
	}

	fmt.Fprint(w, "<br>Your cart:\n")
	for product, quantity := range userCart {
		fmt.Fprintf(w, "%s: %d\n", product, quantity)
	}

	fmt.Fprint(w, "<br>Your Balances: ", balances[username])

	fmt.Fprint(w, `<br><a href="/checkout">Checkout</a>`)

	if minus {
		fmt.Fprint(w, "<br>Your balances is not enough, don't checkout anymore.")
	}
}

func addToCartHandler(w http.ResponseWriter, r *http.Request) {
	product := r.URL.Path[len("/add_to_cart/"):]
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	username := cookie.Value
	userCart, ok := cart[username]
	if !ok {
		userCart = make(map[string]int)
		cart[username] = userCart
	}

	userCart[product]++
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	username := cookie.Value
	userCart, ok := cart[username]
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	total := 0.0
	for product, quantity := range userCart {
		price, ok := products[product]
		if !ok {
			continue
		}
		total += price * float64(quantity)
	}

	if balances[username]-total < 0 {
		minus = true
	} else {
		minus = false
		balances[username] -= total
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
