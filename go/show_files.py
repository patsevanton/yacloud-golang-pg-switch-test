import os

def print_file_contents():
    # Получаем текущую директорию
    current_dir = os.getcwd()

    # Перебираем все элементы в текущей директории
    for item in os.listdir(current_dir):
        # Формируем полный путь к элементу
        item_path = os.path.join(current_dir, item)

        # Проверяем, является ли элемент файлом
        if os.path.isfile(item_path):
            print(f"Файл: {item}")
            print("Содержимое:")
            print("-" * 40)

            try:
                # Открываем файл и читаем его содержимое
                with open(item_path, 'r', encoding='utf-8') as file:
                    content = file.read()
                    print(content)
            except UnicodeDecodeError:
                print("(бинарный файл или нечитаемое содержимое)")
            except Exception as e:
                print(f"Ошибка при чтении файла: {e}")

            print("-" * 40)
            print()  # Пустая строка для разделения

if __name__ == "__main__":
    print_file_contents()
