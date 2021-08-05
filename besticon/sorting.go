package besticon

import "sort"

func sortIcons(icons []Icon, sizeDescending bool) {
	// Order after sorting: (width/height, bytes, url)
	sort.Stable(byURL(icons))
	sort.Stable(byBytes(icons))

	if sizeDescending {
		sort.Stable(sort.Reverse(byWidthHeight(icons)))
	} else {
		sort.Stable(byWidthHeight(icons))
	}
}

type byWidthHeight []Icon

func (a byWidthHeight) Len() int      { return len(a) }
func (a byWidthHeight) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byWidthHeight) Less(i, j int) bool {
	return (a[i].Width < a[j].Width) || (a[i].Height < a[j].Height)
}

type byBytes []Icon

func (a byBytes) Len() int           { return len(a) }
func (a byBytes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byBytes) Less(i, j int) bool { return (a[i].Bytes < a[j].Bytes) }

type byURL []Icon

func (a byURL) Len() int           { return len(a) }
func (a byURL) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byURL) Less(i, j int) bool { return (a[i].URL < a[j].URL) }

func sortIconsByCustom(icons []Icon) {
	sort.Stable(byCustom(icons))
}

type byCustom []Icon

func (list byCustom) Len() int      { return len(list) }
func (list byCustom) Swap(i, j int) { list[i], list[j] = list[j], list[i] }
func (list byCustom) Less(i, j int) bool {
	if list[i].Format < list[j].Format {
		// ico < j** < p**
		return true
	}

	if len(list[i].URL) < len(list[j].URL) {
		return true
	}

	if list[i].Height == 32 {
		return true
	} else if list[j].Height == 32 {
		return false
	}

	if list[i].Height < list[j].Height {
		return true
	}

	return false
}
