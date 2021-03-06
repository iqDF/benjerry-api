package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"

	validatorLib "github.com/iqdf/benjerry-service/common/validator"
	"github.com/iqdf/benjerry-service/domain"
)

// productSingleResponse ...
type productSingleResponse struct {
	Data productResponseData `json:"product"`
}

type productResponseData struct {
	ProductID            string    `json:"productId"`
	Name                 string    `json:"name"`
	ImageClosedURL       string    `json:"image_closed"`
	ImageOpenURL         string    `json:"image_open"`
	Description          string    `json:"description"`
	Story                string    `json:"story"`
	SourcingValues       *[]string `json:"sourcing_values,omitempty"`
	Ingredients          *[]string `json:"ingredients,omitempty"`
	AllergyInfo          string    `json:"allergy_info"`
	DietaryCertification string    `json:"dietary_certifications"`
}

// messageError ....
type messageError struct {
	Message string `json:"message"`
}

type productCreateRequest struct {
	ProductID            string    `json:"productId" validate:"required,numeric,min=3"`
	Name                 string    `json:"name" validate:"required,ascii,max=50"`
	ImageClosedURL       string    `json:"image_closed" validate:"omitempty,uri"`
	ImageOpenURL         string    `json:"image_open" validate:"omitempty,uri"`
	Description          string    `json:"description" validate:"required,max=100"`
	Story                string    `json:"story" validate:"omitempty,max=300"`
	SourcingValues       *[]string `json:"sourcing_values"`
	Ingredients          *[]string `json:"ingredients"`
	AllergyInfo          string    `json:"allergy_info" validate:"required,max=50"`
	DietaryCertification string    `json:"dietary_certifications" validate:"required,max=25"`
}

type productUpdateRequest struct {
	Name                 string    `json:"name" validate:"omitempty,ascii,max=50"`
	ImageClosedURL       string    `json:"image_closed" validate:"omitempty,uri"`
	ImageOpenURL         string    `json:"image_open" validate:"omitempty,uri"`
	Description          string    `json:"description" validate:"omitempty,max=100"`
	Story                string    `json:"story" validate:"omitempty,max=300"`
	SourcingValues       *[]string `json:"sourcing_values" validate:"omitempty"`
	Ingredients          *[]string `json:"ingredients" validate:"omitempty"`
	AllergyInfo          string    `json:"allergy_info" validate:"omitempty,max=50"`
	DietaryCertification string    `json:"dietary_certifications" validate:"omitempty,max=25"`
}

func createToProduct(requestData productCreateRequest) domain.Product {
	return domain.Product{
		ProductID:            requestData.ProductID,
		Name:                 requestData.Name,
		ImageClosedURL:       requestData.ImageClosedURL,
		ImageOpenURL:         requestData.ImageOpenURL,
		Description:          requestData.Description,
		Story:                requestData.Story,
		SourcingValues:       requestData.SourcingValues,
		Ingredients:          requestData.Ingredients,
		AllergyInfo:          requestData.AllergyInfo,
		DietaryCertification: requestData.DietaryCertification,
	}
}

func updateToProduct(requestData productUpdateRequest) domain.Product {
	return domain.Product{
		Name:                 requestData.Name,
		ImageClosedURL:       requestData.ImageClosedURL,
		ImageOpenURL:         requestData.ImageOpenURL,
		Description:          requestData.Description,
		Story:                requestData.Story,
		SourcingValues:       requestData.SourcingValues,
		Ingredients:          requestData.Ingredients,
		AllergyInfo:          requestData.AllergyInfo,
		DietaryCertification: requestData.DietaryCertification,
	}
}

// ProductHandler ...
type ProductHandler struct {
	service domain.ProductService
}

// NewProductHandler creates new HTTP handler
// for product related request
func NewProductHandler(service domain.ProductService) *ProductHandler {
	handler := &ProductHandler{
		service: service,
	}
	return handler
}

func newSingleResponse(product domain.Product) productSingleResponse {
	productData := productResponseData{
		ProductID:            product.ProductID,
		Name:                 product.Name,
		ImageClosedURL:       product.ImageClosedURL,
		ImageOpenURL:         product.ImageOpenURL,
		Description:          product.Description,
		Story:                product.Story,
		SourcingValues:       product.SourcingValues,
		Ingredients:          product.Ingredients,
		AllergyInfo:          product.AllergyInfo,
		DietaryCertification: product.DietaryCertification,
	}
	return productSingleResponse{Data: productData}
}

// Routes register handle func with the path url
func (handler *ProductHandler) Routes(router *mux.Router, middleware alice.Chain) {
	// Register middleware here
	getHandler := middleware.Then(handler.handleGetProduct())
	updateHandler := middleware.Then(handler.handleUpdateProduct())
	deleteHandler := middleware.Then(handler.handleDeleteProduct())
	createHandler := middleware.Then(handler.handleCreateProduct())

	// Register handler methods to router here...
	router.Handle("/{product_id}", getHandler).Methods("GET").Name("PRODUCT_GET")
	router.Handle("/{product_id}", updateHandler).Methods("PUT").Name("PRODUCT_UPDATE")
	router.Handle("/{product_id}", deleteHandler).Methods("DELETE").Name("PRODUCT_DELETE")
	router.Handle("/", createHandler).Methods("POST").Name("PRODUCT_CREATE")
}

// handleGetProduct provides handler func that gets a product
// [GET] /api/products/:product_id
func (handler *ProductHandler) handleGetProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		params := mux.Vars(r)
		productID := params["product_id"]

		product, err := handler.service.GetProduct(r.Context(), productID)

		if err != nil {
			status := getResponseStatus(err)
			writeErrorMessage(w, err.Error(), status)
			return
		}

		response := newSingleResponse(product)
		json.NewEncoder(w).Encode(response)
	}
}

// handleCreateProduct provides handler func that creates a product
// [POST] /api/product/
func (handler *ProductHandler) handleCreateProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var productCreate productCreateRequest
		if err := validatorLib.DecodeAndValidateJSON(r.Body, &productCreate); err != nil {
			verr, _ := err.(*validatorLib.ValidationError)
			writeErrorMessage(w, verr.Message(), http.StatusBadRequest)
			return
		}

		var product = createToProduct(productCreate)
		err := handler.service.CreateProduct(r.Context(), product)

		if err != nil {
			status := getResponseStatus(err)
			writeErrorMessage(w, err.Error(), status)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

// handleUpdateProduct provides handler func that updates a product
// [PUT] /api/product/:product_id
func (handler *ProductHandler) handleUpdateProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		params := mux.Vars(r)
		productID := params["product_id"]

		var productUpdate productUpdateRequest
		if err := validatorLib.ValidateJSON(r.Body, &productUpdate); err != nil {
			verr, _ := err.(*validatorLib.ValidationError)
			writeErrorMessage(w, verr.Message(), http.StatusBadRequest)
			return
		}

		var product = updateToProduct(productUpdate)
		product.ProductID = productID

		err := handler.service.UpdateProduct(r.Context(), productID, product)

		if err != nil {
			status := getResponseStatus(err)
			writeErrorMessage(w, err.Error(), status)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// handleDeleteProduct provides handler func that deletes a product
// [DEL] /api/product/:product_id
func (handler *ProductHandler) handleDeleteProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		params := mux.Vars(r)
		productID := params["product_id"]

		err := handler.service.DeleteProduct(r.Context(), productID)

		if err != nil {
			status := getResponseStatus(err)
			writeErrorMessage(w, err.Error(), status)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// writerErrorMessage is a helper that writes error message to response
func writeErrorMessage(writer http.ResponseWriter, errMsg string, httpStatus int) {
	writer.WriteHeader(httpStatus)
	json.NewEncoder(writer).
		Encode(messageError{Message: errMsg})
}

// getResponseStatus inputs error from application
// and infers the appropriate HTTP status to be returned
func getResponseStatus(err error) int {
	switch err {
	case nil:
		return http.StatusOK
	case domain.ErrAuthFail, domain.ErrExpiredToken:
		return http.StatusUnauthorized
	case domain.ErrBadParamInput:
		return http.StatusBadRequest
	case domain.ErrConflict:
		return http.StatusOK
	case domain.ErrResourceNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
