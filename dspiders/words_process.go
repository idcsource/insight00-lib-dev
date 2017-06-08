// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// The source code is governed by GNU LGPL v3 license

package dspiders

import (
	"fmt"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/drule2/reladb"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// The words index process
type WordsIndexProcess struct {
	queue  chan *WordsIndexRequest // the queue
	count  uint                    // the queue's count
	closed bool                    // if closed it will true

	drule *operator.Operator // The DRule2 remote operator

	sentencedb *reladb.RelaDB // The pages sentence index db

	worddb     *operator.Operator // The word index db
	worddbname string             // The word index db area name

	keyworddb   *operator.Operator // The key word index db
	keywordname string             // The key word index area name
}

// return the index wait queue
func (w *WordsIndexProcess) ReturnQueue() chan *WordsIndexRequest {
	return w.queue
}

// Add a url basic information to the url crawl queue
func (w *WordsIndexProcess) Add(req *WordsIndexRequest) (err error) {
	if w.count == URL_CRAWL_QUEUE_CAP {
		err = fmt.Errorf("The queue is full.")
		return
	}
	w.queue <- req
	w.count++
	return
}

// go to index
func (w *WordsIndexProcess) goindex() {
	for {
		if w.closed == true {
			return
		}

		req := <-w.queue
		switch req.Type {
		case WORDS_INDEX_TYPE_PAGE:
			// if is the page
			w.indexPage(req)
		case WORDS_INDEX_TYPE_AROUND:
			// if is the around link
			w.indexAroundLink(req)
		default:
			continue
		}
	}
}

// the page index
func (w *WordsIndexProcess) indexPage(req *WordsIndexRequest) {
	// TODO operate the KeyWords
	// operate the BodyContent
	word_split := toSequence(req.PageData.BodyContent)
	page_sentences := &PageSentences{
		Url:       req.PageData.Url,
		Sentences: word_split,
	}
	page_sentences.New(req.PageData.Url)
	exist, err := w.sentencedb.TableExist(req.PageData.Url)
	if err != nil {
		fmt.Println(err)
		return
	}
	if exist == false {
		err = w.sentencedb.NewTable(req.PageData.Url, &PageData{}, "Ver")
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	ver, err := w.sentencedb.Insert(req.PageData.Url, req.PageData)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.toindex(req.PageData.Domain, req.PageData.Url, ver, word_split)
}

// to do the word index
func (w *WordsIndexProcess) toindex(domain, url string, ver uint64, split map[uint64][]string) {
	for count, sentence := range split {
		// this is one sentence
		for i, word := range sentence {
			// the sentence len
			slen := len(sentence)
			// this is one word
			tran, errd := w.worddb.Begin()
			if errd.IsError() != nil {
				fmt.Println(errd.IsError())
				break
			}
			// check if the word already exist
			exist, errd := tran.ExistRole(w.worddbname, word)
			if errd.IsError() != nil {
				tran.Rollback()
				fmt.Println(errd.IsError())
				break
			}
			if exist == false {
				// if not exist, create it
				wordindex := &WordIndex{}
				wordindex.New(word)
				errd = tran.StoreRole(w.worddbname, wordindex)
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
			}
			// add the next word index
			if i < slen-1 {
				next_word := sentence[i+1]
				// check the context if exist
				cexist, errd := tran.ExistContext(w.worddbname, word, next_word)
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
				if cexist == false {
					errd = tran.CreateContext(w.worddbname, word, next_word)
					if errd.IsError() != nil {
						tran.Rollback()
						fmt.Println(errd.IsError())
						break
					}
				}
				// the status 0 is the version
				errd = tran.WriteContextStatus(w.worddbname, word, next_word, roles.CONTEXT_DOWN, url, 0, int64(ver))
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
				// the status 1 is the sentence's count
				errd = tran.WriteContextStatus(w.worddbname, word, next_word, roles.CONTEXT_DOWN, url, 1, int64(count))
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
				// the status 2 is the word count in the sentence
				errd = tran.WriteContextStatus(w.worddbname, word, next_word, roles.CONTEXT_DOWN, url, 2, int64(i+1))
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
			}

			// add the prev word index
			if i > 1 {
				prev_word := sentence[i-1]
				// check the context if exist
				cexist, errd := tran.ExistContext(w.worddbname, word, prev_word)
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
				if cexist == false {
					errd = tran.CreateContext(w.worddbname, word, prev_word)
					if errd.IsError() != nil {
						tran.Rollback()
						fmt.Println(errd.IsError())
						break
					}
				}
				// the status 0 is the version
				errd = tran.WriteContextStatus(w.worddbname, word, prev_word, roles.CONTEXT_UP, url, 0, int64(ver))
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
				// the status 1 is the sentence's count
				errd = tran.WriteContextStatus(w.worddbname, word, prev_word, roles.CONTEXT_UP, url, 1, int64(count))
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
				// the status 2 is the word's count in the sentence
				errd = tran.WriteContextStatus(w.worddbname, word, prev_word, roles.CONTEXT_UP, url, 2, int64(i-1))
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
			}

			tran.Commit()
		}
	}
}

// the around link index
func (w *WordsIndexProcess) indexAroundLink(req *WordsIndexRequest) {

}
