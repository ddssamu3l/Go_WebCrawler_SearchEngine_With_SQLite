package main

import "testing"

func TestStop(t *testing.T){
	tests := []struct{
		name string
		words []string
		expected []bool
	}{
		{
			name: "All Stop Words Case",
			words: []string{"'ll","'tis","'twas","'ve","10","39","a","a's",},
			expected: []bool{true, true, true, true, true, true, true, true,},
		},{
			name: "Random words case",
			words: []string{"Apple", "Romeo", "evenly", "Supercar", "furthermore", "furthering", "Earth"},
			expected: []bool{false, false, true, false, true, true, false},
		},
	}

	for _, test := range tests{
		t.Run(test.name, func(t *testing.T) {
			stopwords, err := GenerateStopWords()
			if err != nil{
				t.Errorf("ERROR %s returned: %s\n", test.name, err)
			}

			for index, word := range test.words{
				if test.expected[index] != Stop(word, stopwords){
					t.Errorf("ERROR: %s is cited as not a stop word when it is a stop word.\n", word)
				}
			}
		})
	}
}