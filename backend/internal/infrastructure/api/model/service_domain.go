package model

// ServiceDomainStatus is the status of a domain as the service API shows it.
type ServiceDomainStatus struct {
	State          string         `json:"state" doc:"pending, ready or error. Empty until a controller reported a status for this domain."`
	Reason         string         `json:"reason"`
	Message        string         `json:"message"`
	Records        []DomainRecord `json:"records"`
	ObservedDomain string         `json:"observed_domain" doc:"The domain this status was read for."`
}

// ServiceDomain is one shelf domain, the spec an external controller reconciles.
type ServiceDomain struct {
	Domain    string              `json:"domain"`
	ShelfId   string              `json:"shelf_id"`
	OwnerId   string              `json:"owner_id"`
	CreatedAt string              `json:"created_at" format:"date-time"`
	UpdatedAt string              `json:"updated_at" format:"date-time" doc:"The last domain change or status write."`
	Status    ServiceDomainStatus `json:"status"`
}

// ServiceDomainList is one page of ServiceDomain.
type ServiceDomainList struct {
	Items      []ServiceDomain `json:"items"`
	NextCursor *string         `json:"next_cursor" nullable:"true" doc:"Pass it as cursor to get the next page, null on the last page."`
}

type ServiceDomainListFilter struct {
	Cursor      string `query:"cursor" doc:"The next_cursor of the previous page."`
	Limit       int    `query:"limit" default:"100" minimum:"1" maximum:"500"`
	IfNoneMatch string `header:"If-None-Match" doc:"The ETag of an earlier response. Answers 304 when nothing changed."`
}

type ServiceDomainListResponse struct {
	ETag string            `header:"ETag"`
	Body ServiceDomainList `json:"body"`
}

type ServiceDomainStatusBody struct {
	ShelfId string         `json:"shelf_id" required:"true" doc:"The shelf the controller read the domain from. Answers 409 when that shelf does not have this domain anymore."`
	State   string         `json:"state" required:"true" enum:"pending,ready,error"`
	Reason  string         `json:"reason" required:"false" maxLength:"64" doc:"A machine readable code, for example txt_missing. LinkShelf attaches no meaning to it."`
	Message string         `json:"message" required:"false" maxLength:"1024" doc:"Optional free text fallback for reason."`
	Records []DomainRecord `json:"records" required:"false" maxItems:"20" doc:"The DNS records the user has to set."`
}

type ServiceDomainStatusRequest struct {
	Domain string `path:"domain"`
	Body   ServiceDomainStatusBody
}

type ServiceDomainResponse struct {
	Body ServiceDomain `json:"body"`
}
