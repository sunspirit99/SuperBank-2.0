import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestGetAllUser(t *testing.T) {
	req, err := http.NewRequest("GET", "/get", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(GetAllUser)
	handler.ServeHTTP(w, req)
	fmt.Println(req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)

	}
	expected := `[{"id":1,"name":"sunspirit","balance":90000,"created_time":"2021-07-28 16:40:00","modified_time":"12:00:00"},{"id":11111,"name":"linh","balance":90000,"created_time":"2021-07-28 16:40:00","modified_time":"12:00:00"},{"id":121212,"name":"peter","balance":90000,"created_time":"12:00:00","modified_time":"12:00:00"},{"id":123123,"name":"jack","balance":120000,"created_time":"12:00:00","modified_time":"12:00:00"},{"id":12345,"name":"jack","balance":146000,"created_time":"12:00:00","modified_time":"2021-07-30 10:36:33"},{"id":123456,"name":"jack","balance":102000,"created_time":"12:00:00","modified_time":"2021-07-30 10:36:33"}]`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("handler returned unexpected bot: got %v want %v", w.Body.String(), expected)
	}
}

func TestGetUserById(t *testing.T) {

	req, err := http.NewRequest("GET", "/get/id", nil)
	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{"id": "121212"})
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(GetUserByID)
	handler.ServeHTTP(w, req)
	fmt.Println(req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)

	}

	expected := `{"id":121212,"name":"peter","balance":90000,"created_time":"12:00:00","modified_time":"2021-08-04 15:00:26"}`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("handler returned unexpected bot: got %v want %v", w.Body.String(), expected)
	}
}

func TestUpdateUserById(t *testing.T) {
	var jsonStr = []byte(`{"id":11111,"name":"linh","balance":90000,"created_time":"2021-07-28 16:40:00","modified_time":"12:00:00"}`)
	req, err := http.NewRequest("PUT", "/update", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(UpdateUserByID)
	handler.ServeHTTP(w, req)
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)

	}
	expected := `{"id":11111,"name":"linh","balance":90000,"created_time":"2021-07-28 16:40:00","modified_time":"12:00:00"}`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("handler returned unexpected bot: got %v want %v", w.Body.String(), expected)
	}
}

func TestCreateUser(t *testing.T) {

	var jsonStr = []byte(`{"id":4,"name":"oanh","balance":190000,"created_time":"12:00:00","modified_time":"12:00:00"}`)

	req, err := http.NewRequest("POST", "/create", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateUser)
	handler.ServeHTTP(w, req)
	if status := w.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `{"id":4,"name":"oanh","balance":190000,"created_time":"12:00:00","modified_time":"12:00:00"}`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			w.Body.String(), expected)
	}
}

func TestDeleteUserByID(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/delete/id", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{"id": "121212"})
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(DeleteUserByID)
	handler.ServeHTTP(w, req)
	if status := w.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}
	// expected := `{"id":3,"name":"hung","balance":190000,"created_time":"12:00:00","modified_time":"12:00:00"}`
	// if w.Body.String() != expected {
	// 	t.Errorf("handler returned unexpected body: got %v want %v",
	// 		w.Body.String(), expected)
	// }
}

func TestUserTransfer(t *testing.T) {
	var jsonStr = []byte(`{"id":1, "name":"linh", "amount":5000, "targetID":11111}`)
	req, err := http.NewRequest("PUT", "/transfer", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(UserTransfer)
	handler.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `[{"id":1,"name":"linh","balance":88000,"created_time":"12:00:00","modified_time":"2021-08-04 15:06:11"},{"id":11111,"name":"linh","balance":906000,"created_time":"12:00:00","modified_time":"2021-08-04 15:06:11"}]`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			w.Body.String(), expected)
	}
}
func TestUserWithdraw(t *testing.T) {
	var jsonStr = []byte(`{"id":1, "name":"linh", "amount":5000, "targetID":1}`)
	req, err := http.NewRequest("PUT", "/withdraw", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(UserWithdraw)
	handler.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"id":1,"name":"linh","balance":83000,"created_time":"12:00:00","modified_time":"2021-08-04 15:06:11"},{"id":11111,"name":"linh","balance":906000,"created_time":"12:00:00","modified_time":"2021-08-04 15:06:11"}`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			w.Body.String(), expected)
	}
}

func TestUserDeposit(t *testing.T) {
	var jsonStr = []byte(`{"id":1, "name":"linh", "amount":12000, "targetID":1}`)
	req, err := http.NewRequest("PUT", "/deposit", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(UserDeposit)
	handler.ServeHTTP(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `[{"id":1,"name":"linh","balance":110000,"created_time":"12:00:00","modified_time":"2021-08-04 15:06:11"},{"id":11111,"name":"linh","balance":906000,"created_time":"12:00:00","modified_time":"2021-08-04 15:06:11"}]`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			w.Body.String(), expected)
	}
}
