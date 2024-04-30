class Solution:
    def wordPattern(self, pattern: str, s: str) -> bool:
        words = s.split()

        #pの文字数とsの単語数が違ったら即アウト
        if len(pattern) != len(words):
            return False

        #一対一対応の確認のため、双方向に辞書型
        char_to_word = {}
        word_to_char = {}

        #計算量を考えて、inは不使用
        for char, word in zip(pattern, words):
            mapped_word = char_to_word.get(char)
            mapped_char = word_to_char.get(word)
        
            if mapped_word is None and mapped_char is None:
                char_to_word[char] = word
                word_to_char[word] = char
            elif mapped_word != word or mapped_char != char:
                return False

        return True
    