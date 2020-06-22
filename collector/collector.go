package collector

func setIfSuccess(dst *string, value string) {
	if value != "" {
		*dst = value
	}
}
