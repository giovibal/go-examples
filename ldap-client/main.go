package main

import (
	"gopkg.in/ldap.v2"
	"log"
	"fmt"
	"flag"
)

/**
# DOM_FILI DCs
192.168.21.10	srvw3kdc01
192.168.21.11	srvw3kdc02
192.168.0.15    srvw3kdc03
192.168.0.16    srvw3kdc04


host: srvw3kdc02
port: 389
user: DOM_FILI\root
pass: jaco
dn base: dc=filippetti,dc=it
login: sAMAccountName
 */

//var bindhost = flag.String("bindhost", "192.168.0.15:389", "Bind host:port")
//var binduser = flag.String("binduser", "CN=root,CN=Users,DC=filippetti,DC=it", "Bind username")
//var bindpass = flag.String("bindpass", "jaco", "Bind password")
//var search = flag.String("query", "(cn=root)", "LDAP Search Query")

var bindhost = flag.String("bindhost", "localhost:10389", "Bind host:port")
var binduser = flag.String("binduser", "", "Bind username")
var bindpass = flag.String("bindpass", "", "Bind password")
var querytemplate = flag.String("querytemplate", "(objectClass=%s)", "LDAP Search Query")
var user = flag.String("user", "", "LDAP Username")
var password = flag.String("password", "", "LDAP Password")

func main()  {
	flag.Parse()

	l, err := ldap.Dial("tcp", *bindhost)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	if *binduser != "" && *bindpass != "" {
		err = l.Bind(*binduser, *bindpass)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Pre-Auth %s!\n", "success")
	}

	querystr := fmt.Sprintf(*querytemplate, *user)
	fmt.Printf("Query: %s\n", querystr)

	// SEARCH
	query := ldap.NewSearchRequest(
		"",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		querystr,
		[]string{"cn", "sAMAccountName", "uid"},
		[]ldap.Control{ldap.NewControlPaging(10)},
	)
	result, err := l.Search(query)
	if err != nil {
		log.Fatal(err)
	}
	for i:=0; i<len(result.Entries); i++ {
		e := result.Entries[i]
		fmt.Printf("-> %s\n", e.DN)
		//for a:=0; a<len(e.Attributes); a++ {
		//	attr := e.Attributes[a]
		//	fmt.Printf("----> %v: %v\n", attr.Name, attr.Values)
		//}
		//fmt.Println()
	}

	// if only 1, take it and do a bind
	if len(result.Entries) == 1 {
		user := result.Entries[0]
		err = l.Bind(user.DN, *password)
		if err != nil {
			//log.Fatal(err)
			fmt.Printf("Auth %s!\n", "failed")
		} else {
			fmt.Printf("Auth %s!\n", "success")
		}
	}
}