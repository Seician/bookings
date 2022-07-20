package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	request := httptest.NewRequest("POST", "/whatever", nil)
	form := New(request.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("got invalid when should have been valid")
	}
}
func TestForm_Required(t *testing.T) {
	request := httptest.NewRequest("POST", "/whatever", nil)
	form := New(request.PostForm)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required fields missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "a")
	postedData.Add("c", "a")

	request, _ = http.NewRequest("POST", "/whatever", nil)
	request.PostForm = postedData
	form = New(request.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("form does not have required fields when it does")
	}
}

func TestForm_Has(t *testing.T) {
	request := httptest.NewRequest("POST", "/whatever", nil)
	form := New(request.PostForm)

	has := form.Has("whatever")
	if has {
		t.Error("form shows is has field when it does not")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	form = New(postedData)

	has = form.Has("a")
	if !has {
		t.Error("form does not have field when it should")
	}
}

func TestForm_MinLength(t *testing.T) {
	request := httptest.NewRequest("POST", "/whatever", nil)
	form := New(request.PostForm)

	form.MinLength("a", 3)
	if form.Valid() {
		t.Error("form shows min length for non-existent field")

	}
	isError := form.Errors.Get("a")
	if isError == "" {
		t.Error("should have an error, but did not get one")
	}

	postedData := url.Values{}
	postedData.Add("some_field", "abcd")
	form = New(postedData)

	form.MinLength("some_field", 100)
	if form.Valid() {
		t.Error("shows minlength of 100 met when data is shorter")
	}

	postedData = url.Values{}
	postedData.Add("another_field", "abcd234")
	form = New(postedData)
	form.MinLength("another_field", 2)
	if !form.Valid() {
		t.Error("form does not have min length when it should")
	}

	isError = form.Errors.Get("some_field")
	if isError != "" {
		t.Error("should not have an error, but got one")
	}
}

func TestForm_IsEmail(t *testing.T) {
	postedData := url.Values{}
	form := New(postedData)

	form.IsEmail("x")
	if form.Valid() {
		t.Error("shows email is valid for non-existent field")
	}

	postedData = url.Values{}
	postedData.Add("email", "seician")
	form = New(postedData)

	form.IsEmail("email")
	if form.Valid() {
		t.Error("shows email is valid for invalid email")
	}

	postedData = url.Values{}
	postedData.Add("email", "seician@yahoo.com")

	form = New(postedData)
	if !form.Valid() {
		t.Error("shows email is invalid for valid email")
	}
}
