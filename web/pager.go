// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"math"
	"net/url"
	"strconv"
)

type pager struct {
	uri      *url.URL
	total    int
	pageSize int
	page     int // Pages are 1 based
}

func newPager(uri *url.URL, pageSize int) pager {
	p := pager{
		pageSize: pageSize,
		uri:      uri,
	}
	page, err := strconv.Atoi(uri.Query().Get("page"))
	if err != nil {
		p.page = 1
	}
	p.page = page
	return p
}

// SetTotal sets the total number of records to page
func (p *pager) SetTotal(total int) {
	p.total = total
	if p.Page() > p.Pages() {
		p.page = p.Pages()
	}
}

// Total returns the total number of records
func (p *pager) Total() int {
	return p.total
}

// PageSize returns the current number of records per page
func (p *pager) PageSize() int {
	return p.pageSize
}

// Offset returns the needed query offset for the given page
func (p *pager) Offset() int {
	return (p.Page() - 1) * p.pageSize
}

// Page returns the current page
func (p *pager) Page() int {
	if p.page <= 0 {
		p.page = 1
	}
	return p.page
}

// Pages returns the number of pages in the pager
func (p *pager) Pages() int {
	return int(math.Ceil(float64(p.Total()) / float64(p.pageSize)))
}

// Previous returns the previous page number or -1 if there is no previous page
func (p *pager) Previous() int {
	if p.Page() == 1 {
		return -1
	}
	return p.Page() - 1
}

// Next returns the next page number or -1 if there is no next page
func (p *pager) Next() int {
	if p.Page() == p.Pages() {
		return -1
	}
	return p.Page() + 1
}

// Link creates a link for the given page number
func (p *pager) Link(page int) string {
	values := p.uri.Query()
	values.Set("page", strconv.Itoa(page))
	return "?" + values.Encode()
}

// PageDisplay returns the page links to display, and -1 to display a spacer
func (p *pager) PageDisplay(maxLinks int) []int {
	if maxLinks <= 0 {
		maxLinks = 5
	}

	spacerReserve := 2
	display := make([]int, 0, maxLinks+spacerReserve)

	if maxLinks > p.Pages() {
		for i := 1; i <= p.Pages(); i++ {
			display = append(display, i)
		}
		return display
	}

	firstPageReserve := 1
	lastPageReserve := 1

	low := p.Page() - ((maxLinks / 2) - firstPageReserve)
	high := p.Page() + ((maxLinks / 2) - lastPageReserve)

	if low <= 1 {
		low = firstPageReserve + 1
	}
	if high >= p.Pages() {
		high = p.Pages() - lastPageReserve
	}

	display = append(display, 1) // first page

	// add a space if low is more than 1 page away from first page
	if low > firstPageReserve+1 {
		display = append(display, -1)
	}

	if low >= (p.Pages()-maxLinks)+lastPageReserve+1 {
		low = (p.Pages() - maxLinks) + lastPageReserve + 1
	}

	if high < maxLinks-firstPageReserve {
		high = maxLinks - firstPageReserve
	}

	for i := low; i <= high; i++ {
		display = append(display, i)
	}

	// add a spacer if high is more than one away from last page
	if high < p.Pages()-1 {
		display = append(display, -1) // add a spacer
	}

	display = append(display, p.Pages()) // last page

	return display
}

// func (p *pager) PageDisplay(maxLinks int) []int {
// 	if maxLinks <= 0 {
// 		maxLinks = 5
// 	}

// 	if maxLinks > p.Total() {
// 		maxLinks = p.Total()
// 	}

// 	if p.Pages() == 2 {
// 		return []int{1, 2}
// 	}

// 	spacerReserve := 2
// 	display := make([]int, 0, maxLinks+spacerReserve)

// 	firstPageReserve := 1
// 	lastPageReserve := 1

// 	low := p.Page() - ((maxLinks / 2) - firstPageReserve)
// 	high := p.Page() + ((maxLinks / 2) - lastPageReserve)

// 	if low <= 1 {
// 		low = firstPageReserve + 1
// 	}
// 	if high >= p.Pages() {
// 		high = p.Pages() - lastPageReserve
// 	}

// 	display = append(display, 1) // first page

// 	// add a space if low is more than 1 page away from first page
// 	if low > firstPageReserve+1 && maxLinks < p.Pages() {
// 		display = append(display, -1)
// 	}

// 	if low >= (p.Pages()-maxLinks)+lastPageReserve+1 && low > 1 {
// 		low = (p.Pages() - maxLinks) + lastPageReserve + 1
// 	}

// 	if high < maxLinks-firstPageReserve {
// 		high = maxLinks - firstPageReserve
// 	}

// 	for i := low; i <= high; i++ {
// 		display = append(display, i)
// 	}

// 	// add a spacer if high is more than one away from last page
// 	if high < p.Pages()-1 && maxLinks > p.Pages() {
// 		display = append(display, -1) // add a spacer
// 	}

// 	display = append(display, p.Pages()) // last page

// 	return display
// }
