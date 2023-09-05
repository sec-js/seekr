// Data Structure stored in the DataBase
//
// The person package is used to define a person entry, the main data structure seekr works with.
package person

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/seekr-osint/seekr/api/email"
	"github.com/seekr-osint/seekr/api/enums"
	"github.com/seekr-osint/seekr/api/phone"
	"github.com/seekr-osint/seekr/api/services"
	"github.com/seekr-osint/seekr/api/types/clubs"
	"github.com/seekr-osint/seekr/api/types/hobbies"
	"github.com/seekr-osint/seekr/api/types/ips"
	"github.com/seekr-osint/seekr/api/types/sources"

	"github.com/seekr-osint/seekr/api/validate"
	"gorm.io/gorm"
)

// Person type representing a single person.
type Person struct {
	gorm.Model            `json:"-" tstype:"-" skip:"yes"`
	ID                    uint                           `json:"id" gorm:"primaryKey"` // maybe exploit to overwrite other users data
	Owner                 string                         `json:"-" tstype:"-" validate:"alphanum" skip:"yes"`
	Name                  string                         `json:"name" validate:"required" tstype:",required" example:"john"`
	Age                   uint                           `json:"age" validate:"number" tstype:"number" example:"13"`
	Maidenname            string                         `json:"maidenname" tstype:"string" example:"greg"`
	Kids                  string                         `json:"kids" tstype:"string" example:"no because no wife"`
	Birthday              string                         `json:"bday" tstype:"string" example:"01.01.2001"`
	Address               string                         `json:"address" tstype:"string"`
	Occupation            string                         `json:"occupation" tstype:"string"`
	Prevoccupation        string                         `json:"prevoccupation" tstype:"string"`
	Education             string                         `json:"education" tstype:"string"`
	Military              string                         `json:"military" tstype:"string"`
	Pets                  string                         `json:"pets" tstype:"string"`
	Legal                 string                         `json:"legal" tstype:"string"`
	Political             string                         `json:"political" tstype:"string"`
	Notes                 string                         `json:"notes" tstype:"string"`
	Services              services.MapServiceCheckResult `json:"accounts" grom:"embedded" tstype:"services.MapServiceCheckResult"`
	PhoneNumbers          phone.PhoneNumbers             `json:"phone" tstype:"phone.PhoneNumbers"`
	Emails                email.Emails                   `json:"email" tstype:"email.Emails"`
	Hobbies               hobbies.Hobbies                `json:"hobbies" tstype:"hobbies.Hobbies"`
	Clubs                 clubs.Clubs                    `json:"clubs" tstype:"clubs.Clubs"`
	Sources               sources.Sources                `json:"sources" tstype:"sources.Sources"`
	IPs                   ips.IPs                        `json:"ips" tstype:"ips.IPs"`
	enums.GenderEnum      `json:"gender" tstype:"enums.GenderEnum"`
	enums.EthnicityEnum   `json:"ethnicity" tstype:"enums.EthnicityEnum"`
	enums.CivilstatusEnum `json:"civilstatus" tstype:"enums.CivilstatusEnum"`
	enums.ReligionEnum    `json:"religion" tstype:"enums.ReligionEnum"`
}

// Validate a person and return an error if any field is fornatted incorrectly
// Can be called to make sure a person is valid.
// The data base validates a person on save but this validation is more advanced and detects more errors.
// It uses the go-playground/validator package and struct tags to validate the person.
func (p Person) Validate(personValidator *validate.XValidator) error {
	if personValidator == nil {
		personValidator = NewValidator()
	}
	if errs := personValidator.Validate(p); len(errs) > 0 && errs[0].Error {
		errMsgs := make([]string, 0)

		for _, err := range errs {
			errMsgs = append(errMsgs, fmt.Sprintf(
				"[%s]: '%v' | Needs to implement '%s'",
				err.FailedField,
				err.Value,
				err.Tag,
			))
		}

		return &fiber.Error{
			Code:    fiber.ErrBadRequest.Code,
			Message: strings.Join(errMsgs, " and "),
		}
	}
	return nil
}

// Returns a new validator (*validate.XValidator). Used internally by the Validate() method and has no usage outside of the person package.
func NewValidator() *validate.XValidator {
	v := &validate.XValidator{
		Validator: validator.New(),
	}
	// v.Validator.RegisterValidation("enum",ValidateValuer,false)
	return v
}

// func ValidateValuer(field validator.FieldLevel) bool {
// 	if valuer, ok := field.Field().Interface().(driver.Valuer); ok {

// 		_, err := valuer.Value()
// 		return err == nil
// 	}

// 	return false
// }
