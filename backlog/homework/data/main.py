import openpyxl
import json
import os
import sqlite3
from time import sleep
from deep_translator import GoogleTranslator
import concurrent.futures
from tqdm import tqdm
from argostranslate import package, translate

TABLE_NUM = 1
INDENT = 4

def makeTmpData():
    data = []
    for i in range(1, 4):
        wb = openpyxl.load_workbook(f'base/text_{i}.xlsx')
        sheet = wb.active
        header_row = [cell.value for cell in sheet[1]]
        data_rows = []
        for row in sheet.iter_rows(min_row=2, values_only=True):
            data_rows.append(row)
        for row in data_rows:
            data.append(dict(zip(header_row, row)))

    with open('tmp/mydata.json', 'w') as jsonfile:
        json.dump(data, jsonfile, indent=INDENT)

    labels = {}
    with open('tmp/mydata.json', 'r') as jsonfile:
        data = json.load(jsonfile)
        for row in data:
            if row['label'] in labels:
                labels[row['label']] += 1
            else:
                labels[row['label']] = 0
        print(labels, sep='\n')
        for key in labels:
            with open(f'final/{key}.json', 'w') as f:
                data_rows = []
                for row in data:
                    if row['label'] == key:
                        data_rows.append(row)
                json.dump(data_rows, f, indent=INDENT)
                print(f'done {key}')
    os.remove('tmp/mydata.json')

def fillDb():
    conn = sqlite3.connect('db/mydatabase.db')
    c = conn.cursor()
    # Create table
    c.execute('''CREATE TABLE IF NOT EXISTS emotions
                   (id INTEGER PRIMARY KEY, label TEXT, text TEXT, translate TEXT)''')

    # Insert data from JSON files
    for filename in os.listdir('final'):
        if filename.endswith(".json"):
            with open(f'final/{filename}', 'r') as jsonfile:
                data = json.load(jsonfile)
                for row in data:
                    c.execute("INSERT INTO emotions (label, text, translate) VALUES (?, ?, ?)",
                              (row['label'], row['text'], row['translate']))

    # Save (commit) the changes
    conn.commit()

    # Close the connection
    conn.close()

def fillDbMany():
    conn = [sqlite3.connect(f'db/mydatabase_{i}.db') for i in range(TABLE_NUM)]
    c = [connection.cursor() for connection in conn]

    for db in c:
        db.execute('''CREATE TABLE IF NOT EXISTS emotions
                   (id INTEGER PRIMARY KEY, label TEXT, text TEXT, translate TEXT)''')
        
    # Insert data from JSON files
    for filename in os.listdir('final'):
        if filename.endswith(".json"):
            with open(f'final/{filename}', 'r') as jsonfile:
                data = json.load(jsonfile)
                for count, row in enumerate(data):
                    c[count % TABLE_NUM].execute("INSERT INTO emotions (label, text, translate) VALUES (?, ?, ?)",
                              (row['label'], row['text'], row['translate']))
                    
    for connection in conn:
        connection.commit()
        connection.close()

def translateMarin(db : str, marian_en_ru):
    # from transformers import MarianMTModel, MarianTokenizer
    # from typing import Sequence
    # import torch

    # torch_device = "cuda" if torch.cuda.is_available() else "cpu"

    # print(torch_device)

    # class Translator:
    #     def __init__(self, source_lang: str, dest_lang: str) -> None:
    #         self.model_name = f'Helsinki-NLP/opus-mt-{source_lang}-{dest_lang}'
    #         if torch_device == 'cpu':
    #             self.model = MarianMTModel.from_pretrained(self.model_name)
    #         else:
    #             self.model = MarianMTModel.from_pretrained(self.model_name).to(torch_device)
    #         self.tokenizer = MarianTokenizer.from_pretrained(self.model_name)
            
    #     def translate(self, texts: Sequence[str]) -> Sequence[str]:
    #         tokens = self.tokenizer(list(texts), return_tensors="pt", padding=True)
    #         if torch_device == "cpu":
    #             translate_tokens = self.model.generate(tokens)
    #         else:
    #             translate_tokens = self.model.generate(**tokens.to('cuda'))
    #         return [self.tokenizer.decode(t, skip_special_tokens=True) for t in translate_tokens]
            
    

    # marian_en_ru = Translator('en', 'ru') # Change target language as needed

    conn = sqlite3.connect(db)
    c = conn.cursor()

    while True:
        c.execute("SELECT id, text FROM emotions WHERE translate IS NULL LIMIT 100")
        rows = c.fetchall()
        if not rows:
            break  # Exit the loop if no more rows are found
        for row in rows:
            # print(row)
            id, text = row
            # print(id, text)
            try:
                translated_text = marian_en_ru.translate([text])[0]
                # print(translated_text)
                c.execute("UPDATE emotions SET translate = ? WHERE id = ?", (translated_text, id))
                print(f'Translated id {id}: {text} -> {translated_text}')
            except Exception as e:
                print(f'Error translating id {id}: {e}')
        conn.commit()
    conn.close()

def translate(db : str):
    conn = sqlite3.connect(db)
    c = conn.cursor()
    translator = GoogleTranslator(source='en', target='ru')  # Change target language as needed
    while True:
        c.execute("SELECT id, text FROM emotions WHERE translate IS NULL LIMIT 10")
        rows = c.fetchall()
        if not rows:
            break  # Exit the loop if no more rows are found
        for row in rows:
            id, text = row
            try:
                translated_text = translator.translate(text)
                c.execute("UPDATE emotions SET translate = ? WHERE id = ?", (translated_text, id))
                print(f'Translated id {id}: {text} -> {translated_text}')
            except Exception as e:
                print(f'Error translating id {id}: {e}')
        conn.commit()
        sleep(10)
    conn.close()

def translateMany():

    from transformers import MarianMTModel, MarianTokenizer
    from typing import Sequence
    import torch

    torch_device = "cuda" if torch.cuda.is_available() else "cpu"

    print(torch_device)

    class Translator:
        def __init__(self, source_lang: str, dest_lang: str) -> None:
            self.model_name = f'Helsinki-NLP/opus-mt-{source_lang}-{dest_lang}'
            if torch_device == 'cpu':
                self.model = MarianMTModel.from_pretrained(self.model_name)
            else:
                self.model = MarianMTModel.from_pretrained(self.model_name).to(torch_device)
            self.tokenizer = MarianTokenizer.from_pretrained(self.model_name)
            
        def translate(self, texts: Sequence[str]) -> Sequence[str]:
            tokens = self.tokenizer(list(texts), return_tensors="pt", padding=True)
            if torch_device == "cpu":
                translate_tokens = self.model.generate(tokens)
            else:
                translate_tokens = self.model.generate(**tokens.to('cuda'))
            return [self.tokenizer.decode(t, skip_special_tokens=True) for t in translate_tokens]


    marian_en_ru = Translator('en', 'ru') # Change target language as needed


    with concurrent.futures.ThreadPoolExecutor(max_workers=TABLE_NUM) as executor:
        futures = [executor.submit(translateMarin, f'db/mydatabase_{i}.db', marian_en_ru) for i in range(TABLE_NUM)]
        for future in concurrent.futures.as_completed(futures):
            try:
                future.result()  # add a 30-second timeout
            except concurrent.futures.TimeoutError:
                print(f"Timeout error")

def init_argostranslate():
    os.system('pip3.9 install argostranslate')

    # Download the file
    import urllib.request
    urllib.request.urlretrieve('https://argosopentech.nyc3.digitaloceanspaces.com/argospm/translate-ru_en-1_0.argosmodel', 'translate-ru_en-1_0.argosmodel')

    # Install it
    from argostranslate import package
    package.install_from_path('translate-ru_en-1_0.argosmodel')

def init_MarianMTModel():
    os.system('pip3.11 install pytorch')
    os.system('pip3.11 install transformers')
    os.system('pip3.11 install sentencepiece')
    os.system('pip3.11 install huggingface')

def test_Marin():
    from transformers import MarianMTModel, MarianTokenizer
    from typing import Sequence

    class Translator:
        def __init__(self, source_lang: str, dest_lang: str) -> None:
            self.model_name = f'Helsinki-NLP/opus-mt-{source_lang}-{dest_lang}'
            self.model = MarianMTModel.from_pretrained(self.model_name)
            self.tokenizer = MarianTokenizer.from_pretrained(self.model_name)
            
        def translate(self, texts: Sequence[str]) -> Sequence[str]:
            tokens = self.tokenizer(list(texts), return_tensors="pt", padding=True)
            translate_tokens = self.model.generate(**tokens)
            return [self.tokenizer.decode(t, skip_special_tokens=True) for t in translate_tokens]
            
            
    marian_ru_en = Translator('en', 'ru')
    # res = marian_ru_en.translate(['что слишком сознавать — это болезнь, настоящая, полная болезнь.'])
    res = marian_ru_en.translate(['That being too conscious is a disease, a real, complete disease.'])
    print(res)
    # Returns: ['That being too conscious is a disease, a real, complete disease.']

def parseDoneText():
    insertedRows = 0
    # os.system()
    conn = sqlite3.connect('db/text.db')
    c = conn.cursor()
    c.execute('''CREATE TABLE IF NOT EXISTS emotions
                   (id INTEGER PRIMARY KEY, label TEXT, text TEXT, translate TEXT)''')
    
    tables = [f'db/mydatabase_{i}.db' for i in range(TABLE_NUM)]
    for table in tables:
        conn_ = sqlite3.connect(table)
        c_ = conn_.cursor()
        c_.execute("SELECT text, translate, label FROM emotions WHERE translate IS NOT NULL")
        rows = c_.fetchall()
        conn_.close()
        for row in rows:
            insertedRows += 1
            c.execute("INSERT INTO emotions (label, text, translate) VALUES (?, ?, ?)",
                              (row[2], row[0], row[1]))
    conn.commit()
    conn.close()
    print(f'Rows inserted = {insertedRows}')

def main():
    # cd .\1sem\СТРПО\homework\data\
    # init_MarianMTModel()
    # test_Marin()
    # makeTmpData()
    # fillDb()
    # fillDbMany()
    # translate('db/mydatabase.db')
    translateMany()
    # parseDoneText()

if __name__ == '__main__':
    main()
