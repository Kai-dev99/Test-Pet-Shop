package product

import (
	"bytes"
	"context"
	"errors"
	"go-pet-shop/internal/models"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
)

func requestWithURLParam(req *http.Request, key, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext))
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
}

func TestCreateProduct_BadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(`{"Name":`))

	New(slog.Default(), &ProductsMock{}).CreateProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
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
}

func TestUpdateProduct_BadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/products/7", bytes.NewBufferString(`{"Name":`))
	req = requestWithURLParam(req, "id", "7")

	New(slog.Default(), &ProductsMock{}).UpdateProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
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
}

func TestDeleteProduct_BadRequest(t *testing.T) {
	w := httptest.NewRecorder()

	New(slog.Default(), &ProductsMock{}).DeleteProduct(w, httptest.NewRequest(http.MethodDelete, "/products/", nil))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
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
}
