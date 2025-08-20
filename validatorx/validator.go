package validatorx

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	id_translations "github.com/go-playground/validator/v10/translations/id"
)

var (
	Validate     *validator.Validate
	TranslatorID ut.Translator
	TranslatorEn ut.Translator
)

// InitValidator initializes the global validator instance and sets up
// default translations for both Indonesian and English.
//
// It configures:
//   - A `json` tag lookup function for field name mapping
//   - A global `Validate` instance from `go-playground/validator/v10`
//   - Translators:
//   - `TranslatorID`: Indonesian (`id`)
//   - `TranslatorEn`: English (`en`)
//
// Notes:
//   - If the translator is not found or the translation registration fails,
//     the function will panic.
//
// This function ensures that validation errors use the `json` tag
// instead of the struct field name.
//
// For example:
//
//	Struct field: `Username string `json:"username"`
//	Validation error will show: `username` instead of `Username`
//
// Indonesian Translator:
//   - Uses `id.New()` from go-playground/locales
//   - Registers default translations using `id_translations.RegisterDefaultTranslations`
//
// English Translator:
//   - Uses `en.New()` from go-playground/locales
//   - Registers default translations using `en_translations.RegisterDefaultTranslations`
func InitValidator() {
	Validate = validator.New()
	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tag := fld.Tag.Get("json")
		if tag == "-" {
			return ""
		}

		name := strings.SplitN(tag, ",", 2)[0]
		if name == "" {
			return fld.Name
		}
		return name
	})

	// Setup indonesia translator
	indonesia := id.New()
	uni := ut.New(indonesia, indonesia)

	var found bool
	TranslatorID, found = uni.GetTranslator("id")
	if !found {
		panic("translator not found")
	}

	err := id_translations.RegisterDefaultTranslations(Validate, TranslatorID)
	if err != nil {
		panic(err)
	}

	english := en.New()
	uniEn := ut.New(english, english)

	TranslatorEn, found = uniEn.GetTranslator("en")
	if !found {
		panic("translator not found")
	}

	err = en_translations.RegisterDefaultTranslations(Validate, TranslatorEn)
	if err != nil {
		panic(err)
	}

}
