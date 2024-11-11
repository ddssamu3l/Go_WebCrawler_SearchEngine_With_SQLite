package main

import (
	"reflect"
	"testing"
)

func TestCleanHref(t *testing.T){
	tests := []struct{
		name string
		hostname string
		hrefs []string
		want []string
	}{
		{
			name: "General Case",
			hostname: "https://myusfca.usfca.edu/",
			hrefs: []string{"/top10/", "/dashboard.html", "/dashboard/compsci.html", "/arts-science.html", "https://www.usfca.edu/", "http://myusfca.usfca.edu/"},
			want: []string{"https://myusfca.usfca.edu/top10/", "https://myusfca.usfca.edu/dashboard.html", "https://myusfca.usfca.edu/dashboard/compsci.html", "https://myusfca.usfca.edu/arts-science.html", "INVALID HREF", "http://myusfca.usfca.edu/"},
		},
	}

	for _, test := range tests{
		actualHrefs := Clean(test.hostname, test.hrefs)

		if(!reflect.DeepEqual(test.want, actualHrefs)){
			t.Errorf("ERROR: %s\nExpected: %v\nActual:   %v\n", test.name, test.want, actualHrefs)
		}
	}
}
