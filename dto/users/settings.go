package users

import "strings"

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Jabatan   string `json:"jabatan"`
	Atribut   string `json:"atribut"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func (r *ChangePasswordRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if strings.TrimSpace(r.OldPassword) == "" {
		errors["old_password"] = "password lama harus diisi"
	}
	if len(r.NewPassword) < 8 {
		errors["new_password"] = "password baru minimal 8 karakter"
	}
	if r.NewPassword != r.ConfirmPassword {
		errors["confirm_password"] = "konfirmasi password tidak cocok"
	}
	return errors
}
