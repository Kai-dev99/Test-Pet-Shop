package product

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-pet-shop/internal/models"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-chi/chi"
)

func requestWithURLParam(req *http.Request, key, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext))
}

func assertJSONBody(t *testing.T, w *httptest.ResponseRecorder, expected string) {
	t.Helper()
	var actualBody, expectedBody any
	if err := json.Unmarshal(w.Body.Bytes(), &actualBody); err != nil {
		t.Fatalf("response body is not JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &expectedBody); err != nil {
		t.Fatalf("invalid expected JSON: %v", err)
	}
	if !reflect.DeepEqual(actualBody, expectedBody) {
		t.Errorf("expected body %s, got %s", expected, w.Body.String())
	}
}

func TestGetAllProducts_Success(t *testing.T) {
	mock := &ProductsMock{GetAllProductsFunc: func(context.Context) ([]models.Product, error) {
		return []models.Product{{ID: 1, Name: "Dog Food"}}, nil
	}}
	w := httptest.NewRecorder()

	New(slog.Default(), mock).GetAllProducts(w, httptest.NewRequest(http.MethodGet, "/products", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	assertJSONBody(t, w, `[{"ID":1,"Name":"Dog Food","Price":0,"Stock":0}]`)
}

func TestGetAllProducts_Error(t *testing.T) {
	mock := &ProductsMock{GetAllProductsFunc: func(context.Context) ([]models.Product, error) {
		return nil, errors.New("database error")
	}}
	w := httptest.NewRecorder()

	New(slog.Default(), mock).GetAllProducts(w, httptest.NewRequest(http.MethodGet, "/products", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	assertJSONBody(t, w, `{"error":"Internal server error","message":"Failed to retrieve products"}`)
}

func TestCreateProduct_Success(t *testing.T) {
	expected := models.Product{Name: "Dog Food", Price: 12.5, Stock: 10}
	mock := &ProductsMock{CreateProductFunc: func(_ context.Context, product models.Product) (int, error) {
		if product != expected {
			t.Errorf("expected product %+v, got %+v", expected, product)
		}
		return 1, nil
	}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(`{"Name":"Dog Food","Price":12.5,"Stock":10}`))

	New(slog.Default(), mock).CreateProduct(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	assertJSONBody(t, w, `{"status":"Product created successfully","id":1,"product":{"ID":1,"Name":"Dog Food","Price":12.5,"Stock":10}}`)
}

func TestCreateProduct_BadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(`{"Name":`))

	New(slog.Default(), &ProductsMock{}).CreateProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	assertJSONBody(t, w, `{"error":"Bad request","message":"Invalid JSON payload"}`)
}

func TestCreateProduct_Fail(t *testing.T) {
	mock := &ProductsMock{CreateProductFunc: func(context.Context, models.Product) (int, error) {
		return 0, errors.New("database error")
	}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(`{"Name":"Dog Food","Price":12.5,"Stock":10}`))

	New(slog.Default(), mock).CreateProduct(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	assertJSONBody(t, w, `{"error":"Internal server error","message":"Failed to create product"}`)
}

func TestUpdateProduct_Success(t *testing.T) {
	expected := models.Product{ID: 7, Name: "Dog Food", Price: 12.5, Stock: 10}
	mock := &ProductsMock{UpdateProductFunc: func(_ context.Context, product models.Product) error {
		if product != expected {
			t.Errorf("expected product %+v, got %+v", expected, product)
		}
		return nil
	}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/products/7", bytes.NewBufferString(`{"Name":"Dog Food","Price":12.5,"Stock":10}`))
	req = requestWithURLParam(req, "id", "7")

	New(slog.Default(), mock).UpdateProduct(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	assertJSONBody(t, w, `{"status":"Product updated successfully","id":7,"product":{"ID":7,"Name":"Dog Food","Price":12.5,"Stock":10}}`)
}

func TestUpdateProduct_BadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/products/7", bytes.NewBufferString(`{"Name":`))
	req = requestWithURLParam(req, "id", "7")

	New(slog.Default(), &ProductsMock{}).UpdateProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	assertJSONBody(t, w, `{"error":"Bad request","message":"Invalid JSON payload"}`)
}

func TestUpdateProduct_Fail(t *testing.T) {
	mock := &ProductsMock{UpdateProductFunc: func(context.Context, models.Product) error {
		return errors.New("database error")
	}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/products/7", bytes.NewBufferString(`{"Name":"Dog Food","Price":12.5,"Stock":10}`))
	req = requestWithURLParam(req, "id", "7")

	New(slog.Default(), mock).UpdateProduct(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	assertJSONBody(t, w, `{"error":"Internal server error","message":"Failed to update product"}`)
}

func TestDeleteProduct_Success(t *testing.T) {
	mock := &ProductsMock{DeleteProductFunc: func(_ context.Context, id int) error {
		if id != 7 {
			t.Errorf("expected id 7, got %d", id)
		}
		return nil
	}}
	w := httptest.NewRecorder()
	req := requestWithURLParam(httptest.NewRequest(http.MethodDelete, "/products/7", nil), "id", "7")

	New(slog.Default(), mock).DeleteProduct(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	assertJSONBody(t, w, `{"status":"Product deleted successfully","id":7,"message":"Product with ID 7 has been deleted"}`)
}

func TestDeleteProduct_BadRequest(t *testing.T) {
	w := httptest.NewRecorder()

	New(slog.Default(), &ProductsMock{}).DeleteProduct(w, httptest.NewRequest(http.MethodDelete, "/products/", nil))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	assertJSONBody(t, w, `{"error":"Bad request","message":"Product ID is required"}`)
}

func TestDeleteProduct_Fail(t *testing.T) {
	mock := &ProductsMock{DeleteProductFunc: func(context.Context, int) error {
		return errors.New("database error")
	}}
	w := httptest.NewRecorder()
	req := requestWithURLParam(httptest.NewRequest(http.MethodDelete, "/products/7", nil), "id", "7")

	New(slog.Default(), mock).DeleteProduct(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	assertJSONBody(t, w, `{"error":"Internal server error","message":"Failed to delete product"}`)
}
