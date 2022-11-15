package model

import "wrs/tk/packages/core/license"

func ToLicense(l *license.License) License {
	ret := License{
		ID:   l.LicenseID,
		Name: l.Name,
	}
	if l.GroupID > 0 {
		ret.GroupID = &l.GroupID
		ret.GroupName = &l.Group
	}

	return ret
}
