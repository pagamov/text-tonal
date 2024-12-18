import sqlite3
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import LabelEncoder
from sklearn.metrics import classification_report, accuracy_score
from gensim.models import Word2Vec
import numpy as np
import tensorflow as tf
from tensorflow import keras
from keras import layers, regularizers
import psycopg2
from psycopg2 import sql

from flask import Flask, jsonify

# Database connection parameters
db_params = {
    'dbname': 'database',
    'user': 'pagamov',
    'password': 'multipass',
    'host': 'localhost',  # or your database host
    'port': '5432'        # default PostgreSQL port
}

connection = psycopg2.connect(**db_params)
print("Connection to the database established successfully.")

cursor = connection.cursor()

# Connect to the SQLite database
# conn = sqlite3.connect('db/main.db')
# cursor = conn.cursor()
query = "SELECT id, text_en, text_ru, label, processed FROM sample_table"
cursor.execute(query)
connection.commit()
data = cursor.fetchall()
columns = [description[0] for description in cursor.description]
df = pd.DataFrame(data, columns=columns)
cursor.close()
connection.close()

physical_devices = tf.config.list_physical_devices('GPU')
print("Num GPUs Available: ", len(physical_devices))

# List available GPUs
gpus = tf.config.list_physical_devices('GPU')
if gpus:
    try:
        # Set memory growth to avoid allocating all GPU memory
        for gpu in gpus:
            tf.config.experimental.set_memory_growth(gpu, True)
    except RuntimeError as e:
        print(e)

# Display the first few rows of the dataset
print(df.head())

# Preprocess the text data (tokenization)
df['tokenized_text'] = df['text_en'].apply(lambda x: x.split())

# Train a Word2Vec model
word2vec_model = Word2Vec(sentences=df['tokenized_text'], vector_size=100, window=5, min_count=1, workers=4)
# word2vec_model.

# Function to convert text to vector by averaging word vectors
def text_to_vector(text):
    words = text.split()
    word_vectors = [word2vec_model.wv[word] for word in words if word in word2vec_model.wv]
    if not word_vectors:  # If no words are in the model, return a zero vector
        return np.zeros(word2vec_model.vector_size)
    return np.mean(word_vectors, axis=0)

# Convert the text data to vectors
X = np.array([text_to_vector(text) for text in df['text_en']])
y = df['label']

# Encode labels
label_encoder = LabelEncoder()
y_encoded = label_encoder.fit_transform(y)

# Split the dataset into training and testing sets
X_train, X_test, y_train, y_test = train_test_split(X, y_encoded, test_size=0.2, random_state=42)

# Build the neural network model
model = keras.Sequential([
    layers.Input(shape=(X_train.shape[1],)),  # Input layer
    layers.Dense(256, activation='relu', kernel_regularizer=regularizers.l1(0.01)),       # Hidden layer with 256 neurons
    layers.Dropout(0.5),                        # Dropout layer for regularization
    layers.Dense(128, activation='relu', kernel_regularizer=regularizers.l2(0.01)),        # Hidden layer with 128 neurons
    layers.Dense(len(np.unique(y_encoded)), activation='softmax')  # Output layer
])

# Compile the model
model.compile(optimizer='adam', loss='sparse_categorical_crossentropy', metrics=['accuracy'])

# Train the model
model.fit(X_train, y_train, epochs=10, batch_size=32, validation_split=0.2)  # Use validation split for partitioned training

# Make predictions on the test set
y_pred = model.predict(X_test)
y_pred_classes = np.argmax(y_pred, axis=1)

# Print the classification report
print(classification_report(y_test, y_pred_classes))

# Print the accuracy
accuracy = accuracy_score(y_test, y_pred_classes)
print(f'Accuracy: {accuracy:.2f}')

# Save the model to a file
model.save('text_classification_model.h5')