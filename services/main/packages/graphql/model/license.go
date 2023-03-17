package model

import "wrs/tk/packages/core/license"

func ToLicense(l *license.License) License {
	ret := License{
		Name: l.Name,
	}

	return ret
}
