import os
import zipfile
import hashlib

# program to zip and unzip main db for git storage

db_file = 'db/main.db'  # Path to your database file
zip_file = 'db/main.db.zip'              # Name of the zip file
chunk_size = 10 * 1024 * 1024          # 10 MB

def zip_db_file(db_file_path, zip_file_path):
    with zipfile.ZipFile(zip_file_path, 'w', zipfile.ZIP_DEFLATED) as zipf:
        zipf.write(db_file_path, os.path.basename(db_file_path))
    print(f'delete main db file {db_file_path}')
    os.remove(db_file_path)

def split_file(file_path, chunk_size):
    with open(file_path, 'rb') as f:
        chunk_number = 0
        while True:
            chunk = f.read(chunk_size)
            if not chunk:
                break
            with open(f"{file_path}_part_{chunk_number:03d}", 'wb') as chunk_file:
                chunk_file.write(chunk)
            chunk_number += 1
    print(f'delete db zip file {file_path}')
    os.remove(file_path)

def combine_chunks(chunk_prefix, total_chunks, output_file):
    with open(output_file, 'wb') as outfile:
        for i in range(total_chunks):
            chunk_file_path = f"{chunk_prefix}_part_{i:03d}"
            with open(chunk_file_path, 'rb') as infile:
                outfile.write(infile.read())
            print(f'delete chunk file {i}')
            os.remove(chunk_file_path)

def unzip_file(zip_file_path, extract_to_directory):
    """Unzip a zip file to the specified directory."""
    # Create the directory if it doesn't exist
    os.makedirs(extract_to_directory, exist_ok=True)
    
    with zipfile.ZipFile(zip_file_path, 'r') as zip_ref:
        zip_ref.extractall(extract_to_directory)
        print(f"Extracted all files to '{extract_to_directory}'.")
    print(f'delete db zip file {zip_file_path}')
    os.remove(zip_file_path)

def compare_binary_files(file1_path, file2_path):
    hash1 = hashlib.md5()
    hash2 = hashlib.md5()
    with open(file1_path, 'rb') as file1, open(file2_path, 'rb') as file2:
        chunk_size = 4096
        while True:
            chunk1 = file1.read(chunk_size)
            chunk2 = file2.read(chunk_size)
            if not chunk1 or not chunk2:
                break
            hash1.update(chunk1)
            hash2.update(chunk2)
    if hash1.hexdigest() == hash2.hexdigest():
        print("The files are identical.")
    else:
        print("The files are different.")

def check_file_in_directory(file_name, directory):
    file_path = os.path.join(directory, file_name)
    if os.path.isfile(file_path):
        print(f"The file '{file_name}' exists in the directory '{directory}'.")
        zip_db_file(db_file, zip_file)
        split_file(zip_file, chunk_size)
        return True
    else:
        print(f"The file '{file_name}' does not exist in the directory '{directory}'.")
        combine_chunks(zip_file, len([name for name in os.listdir('./db/') if '_part_' in name]), './db/main.db.zip')
        unzip_file(zip_file, './db/')
        return False
    
check_file_in_directory('main.db', './db/')

# file1 = './db/combined_database.zip'  # Replace with your first file path
# file2 = './db/main.db.zip'  # Replace with your second file path
# compare_binary_files(file1, file2)


