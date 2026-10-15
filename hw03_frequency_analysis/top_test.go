package hw03frequencyanalysis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Change to true if needed.
var taskWithAsteriskIsCompleted = false

var text = `Как видите, он  спускается  по  лестнице  вслед  за  своим
	другом   Кристофером   Робином,   головой   вниз,  пересчитывая
	ступеньки собственным затылком:  бум-бум-бум.  Другого  способа
	сходить  с  лестницы  он  пока  не  знает.  Иногда ему, правда,
		кажется, что можно бы найти какой-то другой способ, если бы  он
	только   мог   на  минутку  перестать  бумкать  и  как  следует
	сосредоточиться. Но увы - сосредоточиться-то ему и некогда.
		Как бы то ни было, вот он уже спустился  и  готов  с  вами
	познакомиться.
	- Винни-Пух. Очень приятно!
		Вас,  вероятно,  удивляет, почему его так странно зовут, а
	если вы знаете английский, то вы удивитесь еще больше.
		Это необыкновенное имя подарил ему Кристофер  Робин.  Надо
	вам  сказать,  что  когда-то Кристофер Робин был знаком с одним
	лебедем на пруду, которого он звал Пухом. Для лебедя  это  было
	очень   подходящее  имя,  потому  что  если  ты  зовешь  лебедя
	громко: "Пу-ух! Пу-ух!"- а он  не  откликается,  то  ты  всегда
	можешь  сделать вид, что ты просто понарошку стрелял; а если ты
	звал его тихо, то все подумают, что ты  просто  подул  себе  на
	нос.  Лебедь  потом  куда-то делся, а имя осталось, и Кристофер
	Робин решил отдать его своему медвежонку, чтобы оно не  пропало
	зря.
		А  Винни - так звали самую лучшую, самую добрую медведицу
	в  зоологическом  саду,  которую  очень-очень  любил  Кристофер
	Робин.  А  она  очень-очень  любила  его. Ее ли назвали Винни в
	честь Пуха, или Пуха назвали в ее честь - теперь уже никто  не
	знает,  даже папа Кристофера Робина. Когда-то он знал, а теперь
	забыл.
		Словом, теперь мишку зовут Винни-Пух, и вы знаете почему.
		Иногда Винни-Пух любит вечерком во что-нибудь поиграть,  а
	иногда,  особенно  когда  папа  дома,  он больше любит тихонько
	посидеть у огня и послушать какую-нибудь интересную сказку.
		В этот вечер...`

func TestTop10(t *testing.T) {
	t.Run("no words in empty string", func(t *testing.T) {
		require.Len(t, Top10(""), 0)
	})

	t.Run("positive test", func(t *testing.T) {
		if taskWithAsteriskIsCompleted {
			expected := []string{
				"а",         // 8
				"он",        // 8
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"в",         // 4
				"его",       // 4
				"если",      // 4
				"кристофер", // 4
				"не",        // 4
			}
			require.Equal(t, expected, Top10(text))
		} else {
			expected := []string{
				"он",        // 8
				"а",         // 6
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"-",         // 4
				"Кристофер", // 4
				"если",      // 4
				"не",        // 4
				"то",        // 4
			}
			require.Equal(t, expected, Top10(text))
		}
	})
}

func TestTop10Cases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single repeated word",
			input:    "cat cat cat",
			expected: []string{"cat"},
		},
		{
			name:     "example from assignment",
			input:    "cat and dog, one dog,two cats and one man",
			expected: []string{"and", "one", "cat", "cats", "dog,", "dog,two", "man"},
		},
		{
			name:     "frequency before lexical order",
			input:    "a b b c c c",
			expected: []string{"c", "b", "a"},
		},
		{
			name:     "equal frequencies",
			input:    "z b a z b a",
			expected: []string{"a", "b", "z"},
		},
		{
			name:     "exactly ten unique words",
			input:    "j i h g f e d c b a",
			expected: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		},
		{
			name:     "fifteen words tied at cutoff",
			input:    "o n m l k j i h g f e d c b a",
			expected: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		},
		{
			name:     "mixed frequencies at cutoff",
			input:    "z z z y y x w v u t s r q p o n m l",
			expected: []string{"z", "y", "l", "m", "n", "o", "p", "q", "r", "s"},
		},
		{
			name:     "unicode whitespace",
			input:    " \tcat\ncat\r\ndog\vbird\fcat\u00a0dog\u2003bird\u2028cat ",
			expected: []string{"cat", "bird", "dog"},
		},
		{
			name:     "unicode words",
			input:    "кот кот 世 世 🙃 ёж ёж",
			expected: []string{"кот", "ёж", "世", "🙃"},
		},
		{
			name:     "case and punctuation are preserved",
			input:    "Нога нога нога, - Нога нога!",
			expected: []string{"Нога", "-", "нога", "нога!", "нога,"},
		},
		{
			name:     "word forms remain distinct",
			input:    "ногу ноги нога ноги",
			expected: []string{"ноги", "нога", "ногу"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, Top10(tc.input))
		})
	}
}

func TestTop10WhitespaceOnly(t *testing.T) {
	require.Empty(t, Top10(" \t\n\r\v\f\u00a0\u2003\u2028"))
}
