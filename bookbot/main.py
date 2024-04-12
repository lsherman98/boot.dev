def main():
    book_path = "books/frankenstein.txt"
    text = get_book_text(book_path)
    num_words = get_num_words(text)

    print('--- Begin report of books/frankenstein.txt ---')
    print(f"{num_words} words found in the document\n")
    counts = count_characters(text)
    for count in counts:
        print(f"The '{count[0]}' character was found {count[1]} times")




def get_num_words(text):
    words = text.split()
    return len(words)


def get_book_text(path):
    with open(path) as f:
        return f.read()


def count_characters(text):
    
    alphabet = 'abcdefghijklmnopqrstuvwxyz'
    counts = {}
    for char in text:
        char = char.lower()
        if char not in alphabet:
            continue
        if char in counts:
            counts[char] += 1
        else:
            counts[char] = 1

    sorted_counts = sorted(counts.items())
    return sorted_counts



main()