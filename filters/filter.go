package filters

import (
	"regexp"
	"strings"

	"github.com/haashemi/tgo"
)

type FilterFunc func(update *tgo.Update) bool

type Filter struct{ f FilterFunc }

func (f Filter) Check(update *tgo.Update) bool { return f.f(update) }

func NewFilter(f FilterFunc) *Filter { return &Filter{f: f} }

// True does nothing and just always returns true.
func True() tgo.Filter {
	return NewFilter(func(update *tgo.Update) bool { return true })
}

// False does nothing and just always returns false.
func False() tgo.Filter {
	return NewFilter(func(update *tgo.Update) bool { return false })
}

// Or behaves like the || operator; returns true if at least one of the passed filters passes.
// returns false if none of them passes.
func Or(filters ...tgo.Filter) tgo.Filter {
	return NewFilter(func(update *tgo.Update) bool {
		for _, filter := range filters {
			if filter.Check(update) {
				return true
			}
		}

		return false
	})
}

// And Behaves like the && operator; returns true if all of the passes filters passes, otherwise returns false.
func And(filters ...tgo.Filter) tgo.Filter {
	return NewFilter(func(update *tgo.Update) bool {
		for _, filter := range filters {
			if !filter.Check(update) {
				return false
			}
		}

		return true
	})
}

// Not Behaves like the ! operator; returns the opposite of the filter result
func Not(filter tgo.Filter) tgo.Filter {
	return NewFilter(func(update *tgo.Update) bool { return !filter.Check(update) })
}

// Text compares the update (message's text or caption, callback query, inline query) with the passed text.
func Text(text string) tgo.Filter {
	return NewFilter(func(update *tgo.Update) bool {
		return ExtractUpdateText(update) == text
	})
}

// Texts compares the update (message's text or caption, callback query, inline query) with the passed texts.
func Texts(texts ...string) tgo.Filter {
	return NewFilter(func(update *tgo.Update) bool {
		raw := ExtractUpdateText(update)

		for _, text := range texts {
			if raw == text {
				return true
			}
		}

		return false
	})
}

// WithPrefix tests whether the update (message's text or caption, callback query, inline query) begins with prefix.
func WithPrefix(prefix string) tgo.Filter {
	return NewFilter(func(update *tgo.Update) bool {
		return strings.HasPrefix(ExtractUpdateText(update), prefix)
	})
}

// WithPrefix tests whether the update (message's text or caption, callback query, inline query) ends with suffix.
func WithSuffix(suffix string) tgo.Filter {
	return NewFilter(func(update *tgo.Update) bool {
		return strings.HasSuffix(ExtractUpdateText(update), suffix)
	})
}

// Regex matches the update (message's text or caption, callback query, inline query) with the passed regexp.
func Regex(reg *regexp.Regexp) tgo.Filter {
	return NewFilter(func(update *tgo.Update) bool {
		return reg.MatchString(ExtractUpdateText(update))
	})
}

// Whitelist compares IDs with the sender-id of the message or callback query. returns true if sender-id is in the blacklist.
func Whitelist(IDs ...int64) tgo.Filter {
	return NewFilter(func(update *tgo.Update) bool {
		var senderID int64

		switch data := ExtractUpdate(update).(type) {
		case *tgo.Message:
			if data.From != nil {
				senderID = data.From.Id
			}
		case *tgo.CallbackQuery:
			senderID = data.From.Id
		default:
			// avoid unnecessary id comparisons.
			return false
		}

		for _, id := range IDs {
			if id == senderID {
				return true
			}
		}

		return false
	})
}

// Blacklist compares IDs with the sender-id of the message or callback query. returns false if sender-id is in the blacklist.
func Blacklist(IDs ...int64) tgo.Filter {
	// Blacklist works the same as Whitelist, So, why not reducing duplicate code!
	return Not(Whitelist(IDs...))
}

func Command(cmd, botUsername string) tgo.Filter { return Commands("/", botUsername, cmd) }

func Commands(prefix, botUsername string, cmds ...string) tgo.Filter {
	// make sure they are all lower-cased
	for index, command := range cmds {
		cmds[index] = strings.ToLower(prefix + command)
	}

	if !strings.HasPrefix(botUsername, "@") {
		botUsername = "@" + botUsername
	}

	return NewFilter(func(update *tgo.Update) bool {
		if msg, ok := ExtractUpdate(update).(*tgo.Message); ok {
			text := msg.Text
			if text == "" {
				text = msg.Caption
			}

			for _, cmd := range cmds {
				if text == cmd || text == cmd+botUsername || strings.HasPrefix(text, cmd+" ") || strings.HasPrefix(text, cmd+botUsername+" ") {
					return true
				}
			}
		}

		// this filter should only work on messages
		return false
	})
}
