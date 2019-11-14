package godig

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

//ClassicDateFormat is used to parse a Salsa timestamp coming from the database.
const ClassicDateFormat = "Mon Jan 2 2006 15:04:05 GMT-0700 (MST)"

//EngageDateFormat is used to format a date for Engage.
const EngageDateFormat = "2006-01-02T15:04:05.000Z"

//DateFormat is used to format a time so that Engage will recognize it.
const DateFormat = "2006-01-02"

//TimestampFormat is used to format a time so that Engage will recognize it.
const TimestampFormat = "2006-01-02T15:04:05"

//API hold the data that we need to do Salsa API calls.  That includes
//the cookies from authentication.
type API struct {
	Client   *http.Client
	Cookies  []*http.Cookie
	Host     string
	Verbose  bool
	CredData CredData
}

//Table links an API to a Salsa database table.
type Table struct {
	*API
	Name string
}

//AuthStatus contains the information returned by Authentication.
type AuthStatus struct {
	Status  string
	Message string
}

//Results is returned by API calls.
type Results struct {
	Object   string   `json:"object"`
	Key      string   `json:"key"`
	Result   string   `json:"result"`
	Messages []string `json:"messages"`
}

//CredData contains the info that we need to get into the API.
type CredData struct {
	Host     string
	Email    string
	Password string
}

//DeleteStatus contins the info returned by deleting a record.
type DeleteStatus struct {
	Object   string
	Key      string
	Result   string
	Messages []string
}

//Field is used to describe table fields when calling Describe.
type Field struct {
	DisplayToSupporters string `json:"display_to_supporters,omitempty"`
	Name                string `json:"name,omitempty"`
	DataColumn          string `json:"data_column,omitempty"`
	IsAZeroIndexEnum    bool   `json:"is_a_zero_index_enum,string,omitempty"`
	Label               string `json:"label,omitempty"`
	DataTable           string `json:"data_table,omitempty"`
	DisplayName         string `json:"displayName,omitempty"`
	Type                string `json:"type,omitempty"`
	IsCustom            bool   `json:"isCustom,omitempty,string"`
}

//FieldList is a slice of Fields returned by Describe.
type FieldList []Field

//MapList is a slice of FieldMaps.
type MapList []gjson.Result

//NewAPI initializes and returns an API object.
func NewAPI() *API {
	c := API{}
	c.Client = &http.Client{}
	return &c
}

//NewTable creates a table using a table/object name.
func (a *API) NewTable(n string) Table {
	t := Table{a, n}
	return t
}

//Donation is a shortcut for creating a donation Table object.
func (a *API) Donation() Table {
	return a.NewTable("donation")
}

//EmailBlast is a shortcut for creating an EmailBlast Table object.
func (a *API) EmailBlast() Table {
	return a.NewTable("email_blast")
}

//Groups is a shorcut for creating a groups Table.  Note that
//"groups" is the only table in the API that's plural.
func (a *API) Groups() Table {
	return a.NewTable("groups")
}

//GroupsSupporters is a shortcut to join groups to supporters
//via the supporter_groups table. Use LeftJoin to get data for
//this object.
func (a *API) GroupsSupporters() Table {
	return a.NewTable("groups(groups_KEY)supporter_groups(supporter_KEY)supporter")
}

//Org is a shortcut for creating an organization Table object.
func (a *API) Org() Table {
	return a.NewTable("organization")
}

//Supporter is a shortcut for creating a supporter Table object.
func (a *API) Supporter() Table {
	return a.NewTable("supporter")
}

//SupporterDonation is a shortcut for creating a Table that
//holds supporter and donation records.  Use LeftJoin to get
//data for this object.
func (a *API) SupporterDonation() Table {
	return a.NewTable("supporter(supporter_KEY)donation")
}

//SupporterGroups is a shortcut for creating a supporter_group Table object.
func (a *API) SupporterGroups() Table {
	return a.NewTable("supporter_groups")
}

//Publish is a shortcut for creating a publish Table object.
func (a *API) Publish() Table {
	return a.NewTable("publish")
}

//ShortDate accepts a time and outputs it as YYYY-mm-dd.
func ShortDate(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return s
	}
	t, err := time.Parse(ClassicDateFormat, s)
	if err != nil {
		log.Fatalf("%v: %v\n", err, s)
	}
	return t.Format(DateFormat)
}

//EngageDate parses converts a string containing a MySQL date to
//another string containing an Engage date.
func EngageDate(s string) string {
	t, err := time.Parse(ClassicDateFormat, s)
	if err != nil {
		log.Printf("Warning: parsing %v returned %v\n", s, err)
	} else {
		s = t.Format(EngageDateFormat)
	}
	return s
}

//EngageTimestamp parses converts a string containing a MySQL date to
//another string containing an Engage date and time.
func EngageTimestamp(s string) string {
	// Date_Created, Transaction_Date, etc.  Convert dates from MySQL to Engage.
	p := strings.Split(s, " ")
	if len(p) >= 7 {
		//Pull out the timezone.
		p = append(p[0:5], p[6])
		x := strings.Join(p, " ")
		t, err := time.Parse(ClassicDateFormat, x)
		if err != nil {
			log.Printf("Warning: parsing %v returned %v\n", s, err)
		} else {
			s = t.Format(TimestampFormat)
		}
	}
	return s
}

//Organization describes a record for an Salsa Classic client.
//Be aware that the bulk of the non-identity and non-status fields
//have been deprecated.
type Organization struct {
	OrganizationKEY                            string `json:"organization_KEY"`
	RootKEY                                    string `json:"root_KEY,omitempty"`
	ParentKEY                                  string `json:"parent_KEY,omitempty"`
	PartnerKEY                                 string `json:"partner_KEY,omitempty"`
	LastModified                               string `json:"Last_Modified,omitempty"`
	DateCreated                                string `json:"Date_Created,omitempty"`
	PRIVATEDateCreated                         string `json:"PRIVATE_Date_Created,omitempty"`
	Name                                       string `json:"Name"`
	Type                                       string `json:"Type"`
	Status                                     string `json:"Status"`
	READONLYShortName                          string `json:"READONLY_Short_Name,omitempty"`
	Description                                string `json:"Description,omitempty"`
	OrganizationHomepage                       string `json:"Organization_Homepage,omitempty"`
	NewsletterOrListserveName                  string `json:"Newsletter_or_Listserve_Name,omitempty"`
	CustomHeaderHTML                           string `json:"Custom_Header_HTML,omitempty"`
	CustomFooterHTML                           string `json:"Custom_Footer_HTML,omitempty"`
	PrintHeader                                string `json:"Print_Header,omitempty"`
	PrintFooter                                string `json:"Print_Footer,omitempty"`
	BaseURL                                    string `json:"Base_URL"`
	SecureURL                                  string `json:"Secure_URL"`
	Street                                     string `json:"Street,omitempty"`
	Street2                                    string `json:"Street_2,omitempty"`
	City                                       string `json:"City,omitempty"`
	State                                      string `json:"State,omitempty"`
	Zip                                        string `json:"Zip,omitempty"`
	PRIVATEZipPlus4                            string `json:"PRIVATE_Zip_Plus_4,omitempty"`
	MailServer                                 string `json:"mail_server,omitempty"`
	MailUser                                   string `json:"mail_user,omitempty"`
	MailPass                                   string `json:"mail_pass,omitempty"`
	MailStatus                                 string `json:"Mail_Status,omitempty"`
	PromotionalCode                            string `json:"Promotional_Code,omitempty"`
	Interests                                  string `json:"Interests,omitempty"`
	Note                                       string `json:"Note,omitempty"`
	GlobalMailHTMLFooter                       string `json:"Global_Mail_HTML_Footer,omitempty"`
	GlobalMailTextFooter                       string `json:"Global_Mail_Text_Footer,omitempty"`
	LinkTrackURL                               string `json:"Link_Track_URL,omitempty"`
	OpenTrackURL                               string `json:"Open_Track_URL,omitempty"`
	SalsifiedBOOLVALUE                         string `json:"salsified_BOOLVALUE,omitempty"`
	Salsified                                  bool   `json:"salsified,omitempty"`
	StatusLastModified                         string `json:"Status_Last_Modified,omitempty"`
	DateTrialStarted                           string `json:"Date_Trial_Started,omitempty"`
	ContractDate                               string `json:"Contract_Date,omitempty"`
	ToolsInContract                            string `json:"Tools_In_Contract,omitempty"`
	ClosedDate                                 string `json:"Closed_Date,omitempty"`
	ClosedReason                               string `json:"Closed_Reason,omitempty"`
	DefaultEmailAddress                        string `json:"default_email_address,omitempty"`
	DefaultMerchantAccountKEY                  string `json:"default_merchant_account_KEY,omitempty"`
	MovedBOOLVALUE                             string `json:"moved_BOOLVALUE,omitempty"`
	Moved                                      bool   `json:"moved,omitempty"`
	BlastNotificationEmail                     string `json:"Blast_Notification_Email,omitempty"`
	Country                                    string `json:"Country,omitempty"`
	ListSize                                   string `json:"List_Size,omitempty"`
	Usages                                     string `json:"Usages,omitempty"`
	HearAboutUs                                string `json:"hear_about_us,omitempty"`
	Tier                                       string `json:"Tier,omitempty"`
	TaxStatus                                  string `json:"Tax_Status,omitempty"`
	StaffContact                               string `json:"Staff_Contact,omitempty"`
	LanguageCode                               string `json:"language_code,omitempty"`
	DisableTokenAuthenticationBOOLVALUE        string `json:"Disable_Token_Authentication_BOOLVALUE,omitempty"`
	DisableTokenAuthentication                 bool   `json:"Disable_Token_Authentication,omitempty"`
	EnforcePackagePermissions                  string `json:"Enforce_Package_Permissions,omitempty"`
	RecommendedMailServer                      string `json:"recommended_mail_server,omitempty"`
	OverrideMailServerBOOLVALUE                string `json:"override_mail_server_BOOLVALUE,omitempty"`
	OverrideMailServer                         bool   `json:"override_mail_server,omitempty"`
	EmailBlastBrandingOptOutBOOLVALUE          string `json:"Email_Blast_Branding_Opt_Out_BOOLVALUE,omitempty"`
	EmailBlastBrandingOptOut                   bool   `json:"Email_Blast_Branding_Opt_Out,omitempty"`
	ExternalClientID                           string `json:"external_client_id,omitempty"`
	RecommendEmailBlastBrandingOptOutBOOLVALUE string `json:"Recommend_Email_Blast_Branding_Opt_Out_BOOLVALUE,omitempty"`
	RecommendEmailBlastBrandingOptOut          bool   `json:"Recommend_Email_Blast_Branding_Opt_Out,omitempty"`
	OrganizationID                             string `json:"organization_id,omitempty"`
	WebsiteForcedBrandingRequiredBOOLVALUE     string `json:"Website_Forced_Branding_Required_BOOLVALUE,omitempty"`
	WebsiteForcedBrandingRequired              bool   `json:"Website_Forced_Branding_Required,omitempty"`
	AuthenticateEmailsBOOLVALUE                string `json:"Authenticate_Emails_BOOLVALUE,omitempty"`
	AuthenticateEmails                         bool   `json:"Authenticate_Emails,omitempty"`
	EnforceAutodedupeOnEmailSendBOOLVALUE      string `json:"enforce_autodedupe_on_email_send_BOOLVALUE,omitempty"`
	EnforceAutodedupeOnEmailSend               bool   `json:"enforce_autodedupe_on_email_send,omitempty"`
	EnforceHTTPSBOOLVALUE                      string `json:"enforce_https_BOOLVALUE,omitempty"`
	EnforceHTTPS                               bool   `json:"enforce_https,omitempty"`
}
