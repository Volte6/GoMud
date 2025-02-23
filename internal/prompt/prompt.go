package prompt

import (
	"strings"
)

/*

// Usage

// At user input submission:

if cmdPrompt := prompt.Get(userId); cmdPrompt != nil {

	for {
		question := cmdPrompt.GetQuestion()

		if question == nil {
			break
		}

		question.Answer(inputString)
		// Display the current prompt, and wait for a response
		print( question.Question )

	}

}

*/

type Question struct {
	Question        string   // What's the prompt?
	Options         []string // What options (if any) are available? None = freeform
	DefaultResponse string   // What is the default response?
	Response        string   // What was the response last submitted
	Done            bool     // Was it seen and responded to?
	Flags           int      // Mask reply etc
}

type Prompt struct {
	Command   string         // Where does it call when complete?
	Rest      string         // What is the 'rest' of the command
	Questions []*Question    // All questions so far
	State     map[string]any // Optional place to store state
}

func New(command string, rest string) *Prompt {
	return &Prompt{
		Command:   command,
		Rest:      rest,
		Questions: make([]*Question, 0),
		State:     map[string]any{},
	}
}

// Returns the next pending question.
func (p *Prompt) Ask(question string, responseOptions []string, defaultOption ...string) *Question {

	if question == `` {
		question = `?`
	}

	qCt := len(p.Questions)
	for i := 0; i < qCt; i++ {
		if p.Questions[i].Question == question {
			return p.Questions[i]
		}
	}

	defOpt := ``
	if len(defaultOption) > 0 {
		defOpt = defaultOption[0]
	}

	finalResponseOptions := []string{}
	for _, opt := range responseOptions {
		if opt != `` {
			finalResponseOptions = append(finalResponseOptions, opt)
		}
	}

	q := &Question{
		Question:        question,
		Options:         finalResponseOptions,
		DefaultResponse: defOpt,
	}

	p.Questions = append(p.Questions, q)

	return q
}

// Returns the next pending question.
func (p *Prompt) GetNextQuestion() *Question {

	qCt := len(p.Questions)
	for i := 0; i < qCt; i++ {
		if !p.Questions[i].Done {
			return p.Questions[len(p.Questions)-1]
		}
	}

	return nil
}

func (p *Prompt) Store(name string, val any) {
	p.State[name] = val
}

func (p *Prompt) Recall(name string) (any, bool) {
	if v, ok := p.State[name]; ok {
		return v, true
	}
	return nil, false
}

func (q *Question) Reset() {
	q.Done = false
}

/*
If fewer than 2 choices provided, answer is what they enter.
If they provide an empty string, and a default respnose is provided,
answer is default.
*/
func (q *Question) Answer(answer string) {
	// If an empty string, failover to default (if any)
	// Otherwise, just abort and wait for a valid response
	answer = strings.TrimSpace(answer)
	if len(answer) == 0 {
		if q.DefaultResponse == `` {
			return
		}
		q.Response = q.DefaultResponse
		q.Done = true
	}

	// If options were provided, find best match if any
	optLen := len(q.Options)
	if optLen > 1 {

		closestMatchIdx := -1
		closestMatchLen := 0

		testAnswer := strings.ToLower(answer)

		for i := 0; i < optLen; i++ {
			optTest := strings.ToLower(q.Options[i])

			if optTest == testAnswer {
				closestMatchIdx = i
				break
			}

			longestPossible := len(optTest)
			if len(testAnswer) < longestPossible {
				longestPossible = len(testAnswer)
			}
			for j := 0; j < longestPossible; j++ {
				if optTest[j] != testAnswer[j] {
					break
				}

				if j+1 > closestMatchLen {
					closestMatchLen = j + 1
					closestMatchIdx = i
				}
			}
		}

		if closestMatchIdx == -1 {
			return
		}

		answer = q.Options[closestMatchIdx]
	}

	if len(answer) > 0 {
		q.Response = answer
	}
	q.Done = true
}

func (q *Question) RejectResponse() {

	q.Response = `` // Clear the response
	q.Done = false  // Mark as not done
}

func (q *Question) String() string {

	ret := strings.Builder{}
	ret.WriteString(`<ansi fg="black-bold">.:</ansi> `) // Prompt prefix
	ret.WriteString(`<ansi fg="yellow-bold">`)
	ret.WriteString(q.Question) // Actual question
	ret.WriteString(`</ansi>`)

	optLen := len(q.Options)
	if optLen > 0 {

		ret.WriteString(` <ansi fg="black-bold">[</ansi>`)
		for i := 0; i < optLen; i++ {

			if q.DefaultResponse != `` {

				if q.Options[i] == q.DefaultResponse {
					ret.WriteString(`<ansi fg="white">`)
				} else {
					ret.WriteString(`<ansi fg="black-bold">`)
				}

				ret.WriteString(q.Options[i])

				ret.WriteString(`</ansi>`)

			} else {
				ret.WriteString(q.Options[i])
			}

			if i < optLen-1 {
				ret.WriteString(`<ansi fg="black-bold">/</ansi>`)
			}
		}
		ret.WriteString(`<ansi fg="black-bold">]</ansi>`)
	}
	ret.WriteString(` `)
	return ret.String()
}
