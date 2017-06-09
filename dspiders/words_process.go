// Copyright 2016-2017
// CoderG the 2016 project
// Insight 0+0 [ 洞悉 0+0 ]
// InDimensions Construct Source [ 忆黛蒙逝·建造源 ] -> idcsource@gmail.com
// Stephen Fire Meditation Qin [ 火志溟 ] -> firemeditation@gmail.com
// The source code is governed by GNU LGPL v3 license

package dspiders

import (
	"fmt"
	"strconv"

	"github.com/idcsource/Insight-0-0-lib/drule2/operator"
	"github.com/idcsource/Insight-0-0-lib/roles"
)

// The words index process
type WordsIndexProcess struct {
	queue  chan *WordsIndexRequest // the queue
	count  uint                    // the queue's count
	closed bool                    // if closed it will true

	drule *operator.Operator // The DRule2 remote operator

	//sentencedb *reladb.RelaDB // The pages sentence index db

	sentencedb     *operator.Operator // The pages sentence index db
	sentencedbname string             // The pages sentence index db area name

	worddb     *operator.Operator // The word index db
	worddbname string             // The word index db area name

	keyworddb   *operator.Operator // The key word index db
	keywordname string             // The key word index area name
}

func NewWordsIndexProcess(sentencedb *operator.Operator, sentencedbname string, worddb *operator.Operator, worddbname string) (w *WordsIndexProcess) {
	w = &WordsIndexProcess{
		queue:          make(chan *WordsIndexRequest, URL_CRAWL_QUEUE_CAP),
		count:          0,
		closed:         true,
		sentencedb:     sentencedb,
		sentencedbname: sentencedbname,
		worddb:         worddb,
		worddbname:     worddbname,
	}
	return
}

// Start the processor
func (w *WordsIndexProcess) Start() {
	w.closed = false
	go w.goindex()
}

// Stop the processor
func (w *WordsIndexProcess) Close() {
	w.closed = true
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
	/*
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
		}*/
	exist, errd := w.sentencedb.ExistRole(w.sentencedbname, req.PageData.Url)
	if errd.IsError() != nil {
		fmt.Println(errd.IsError())
		return
	}
	if exist == true {
		var old_split map[uint64][]string
		errd = w.sentencedb.ReadData(w.sentencedbname, req.PageData.Url, "Sentences", &old_split)
		if errd.IsError() != nil {
			fmt.Println(errd.IsError())
			return
		}
		w.todelindex(req.PageData.Domain, req.PageData.Url, old_split)
	}
	errd = w.sentencedb.StoreRole(w.sentencedbname, page_sentences)
	w.toindex(req.PageData.Domain, req.PageData.Url, word_split)
}

func (w *WordsIndexProcess) todelindex(domain, url string, split map[uint64][]string) {
	for _, sentence := range split {
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
				tran.Rollback()
				break
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
				if cexist == true {
					// read already exist index set, is status 0 string
					var index_set string
					have, errd := tran.ReadContextStatus(w.worddbname, word, next_word, roles.CONTEXT_DOWN, url, 0, &index_set)
					if errd.IsError() != nil {
						tran.Rollback()
						fmt.Println(errd.IsError())
						break
					}
					if have == true {
						errd = tran.DeleteContextBind(w.worddbname, word, next_word, roles.CONTEXT_DOWN, url)
						if errd.IsError() != nil {
							tran.Rollback()
							fmt.Println(errd.IsError())
							break
						}
					}
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
				if cexist == true {
					// read already exist index set, is status 0 string
					var index_set string
					have, errd := tran.ReadContextStatus(w.worddbname, word, prev_word, roles.CONTEXT_UP, url, 0, &index_set)
					if errd.IsError() != nil {
						tran.Rollback()
						fmt.Println(errd.IsError())
						break
					}
					if have == true {
						errd = tran.DeleteContextBind(w.worddbname, word, prev_word, roles.CONTEXT_UP, url)
						if errd.IsError() != nil {
							tran.Rollback()
							fmt.Println(errd.IsError())
							break
						}
					}
				}
			}

			tran.Commit()
		}
	}
}

// to do the word index
func (w *WordsIndexProcess) toindex(domain, url string, split map[uint64][]string) {
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
				// the word's count
				word_count := int64(count) + int64(i+1)
				// read already exist index set, is status 0 string
				var index_set string
				have, errd := tran.ReadContextStatus(w.worddbname, word, next_word, roles.CONTEXT_DOWN, url, 0, &index_set)
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
				if have == false {
					// if not have
					index_set = strconv.FormatInt(word_count, 10)
				} else {
					index_set = index_set + " " + strconv.FormatInt(word_count, 10)
				}
				errd = tran.WriteContextStatus(w.worddbname, word, next_word, roles.CONTEXT_DOWN, url, 0, index_set)
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
				// the word's count
				word_count := int64(count) + int64(i-1)
				// read already exist index set, is status 0 string
				var index_set string
				have, errd := tran.ReadContextStatus(w.worddbname, word, prev_word, roles.CONTEXT_UP, url, 0, &index_set)
				if errd.IsError() != nil {
					tran.Rollback()
					fmt.Println(errd.IsError())
					break
				}
				if have == false {
					// if not have
					index_set = strconv.FormatInt(word_count, 10)
				} else {
					index_set = index_set + " " + strconv.FormatInt(word_count, 10)
				}
				errd = tran.WriteContextStatus(w.worddbname, word, prev_word, roles.CONTEXT_UP, url, 0, index_set)
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
	// TODO
}
