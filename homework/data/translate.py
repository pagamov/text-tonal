import concurrent.futures
import json
import os

from time import sleep
from deep_translator import GoogleTranslator
from tqdm import tqdm

INDENT = 4
labels = ['fear','sadness','love','joy','surprise','anger','neutral','hate','worry','relief','happiness','fun','empty','enthusiasm','boredom']

for label in labels:
    translator = GoogleTranslator(source='en', target='ru')
    data = []
    res_data = []
    new_data = []
    translation_count = [0]  # Using a list to hold the count

    def translate_row(row):
        row["translate"] = translator.translate(row["text"])
        res_data.append(row)
        translation_count[0] += 1  # Increment the count

        # # Save to file every 100 translations
        if translation_count[0] % 100 == 0:
            with open(f'final/{label}_partial_{translation_count[0]}.json', 'w') as f:
                json.dump(res_data, f, indent=INDENT)
                print(f'Saved {translation_count[0]} translations to {label}_partial_{translation_count[0]}.json')

    print(f'### doing {label} ###')
    with open(f'final/{label}.json', 'r') as jsonfile:
        data = json.load(jsonfile)

    for row in data:
        if row["translate"] != None:
            new_data.append(row)
        else:
            res_data.append(row)
    with concurrent.futures.ThreadPoolExecutor(max_workers=16) as executor:
        futures = [executor.submit(translate_row, row) for row in new_data]
        for future in tqdm(concurrent.futures.as_completed(futures), total=len(new_data)):
            try:
                future.result(timeout=5)  # add a 5-second timeout
            except concurrent.futures.TimeoutError:
                print(f"Timeout error")

    # Final save after all translations
    with open(f'final/{label}.json', 'w') as jsonfile:
        json.dump(res_data, jsonfile, indent=INDENT)

    print(f'All translations for {label} completed and saved.')