import os
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
import logging
import joblib
from flask import Flask, jsonify, request

# from flask import Flask

app = Flask(__name__)

main_model = None
w2v_model = None
label_encoder = None

logging.basicConfig(level=logging.DEBUG,  # Set the logging level
                    format='%(asctime)s - %(levelname)s - %(message)s')  # Set the log message format


@app.route('/ping', methods=['GET'])
def ping():
    return jsonify({'message': 'pong'})

# @app.route('/status', methods=['GET'])
# def status():
#     return main_model.summary()

@app.route('/predict', methods=['POST'])
def predict():
    text : str = request.get_json(force=True)["text"]
    print(text)
    vector = text_to_vector(text.lower(), w2v_model).reshape(1,-1)
    res = main_model.predict(vector)
    pred_class_index = np.argmax(res, axis=1)[0]

    by_word = []
    for word in text.split(" "):
        vector = text_to_vector(word.lower(), w2v_model).reshape(1,-1)
        res = main_model.predict(vector)
        # print(res, res.shape)
        pred_class_index = np.argmax(res, axis=1)[0]
        by_word.append({'Word': word, 'Label': label_encoder.classes_[pred_class_index]})
    return jsonify({'Label': label_encoder.classes_[pred_class_index], 'Words': by_word})



def connect() -> tuple[psycopg2.extensions.connection, psycopg2.extensions.cursor]:
    # Connect to the PostgreSQL database
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
    return connection, cursor

def get_data() -> pd.DataFrame:
    connection, cursor = connect()
    query = "SELECT id, text_en, text_ru, label, processed FROM sample_table"
    cursor.execute(query)
    connection.commit()
    data = cursor.fetchall()
    columns = [description[0] for description in cursor.description]
    df = pd.DataFrame(data, columns=columns)
    cursor.close()
    connection.close()
    return df

def check_cuda():
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

# Function to convert text to vector by averaging word vectors
def text_to_vector(text : str, word2vec_model : Word2Vec) -> np.ndarray:
    words = text.split()
    word_vectors = [word2vec_model.wv[word] for word in words if word in word2vec_model.wv]
    if not word_vectors:  # If no words are in the model, return a zero vector
        return np.zeros(word2vec_model.vector_size)
    return np.mean(word_vectors, axis=0)

@app.route('/train')
def train():
    # check_cuda()
    df = get_data()
    print(df.head())
    df['tokenized_text'] = df['text_en'].apply(lambda x: x.split())
    word2vec_model = Word2Vec(sentences=df['tokenized_text'], vector_size=100, window=5, min_count=1, workers=4)

    save_W2V_model(word2vec_model)

    # Convert the text data to vectors
    X = np.array([text_to_vector(text, word2vec_model) for text in df['text_en']])
    y = df['label']

    # Encode labels
    label_encoder = LabelEncoder()
    y_encoded = label_encoder.fit_transform(y)

    joblib.dump(label_encoder, "label_encoder.pkl")

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

    save_model(model)

    return jsonify({'message': 'trained'}) 

def save_model(model: keras.Model):
    model.save('text_classification_model.h5')

def save_W2V_model(model):
    model.save('word2vec_model.model')

def loadWord2VecModel():
    model_path = os.path.join('./', 'word2vec_model.model')
    if os.path.isfile(model_path):
        # Load the model if it exists
        # model = keras.model.load
        model = Word2Vec.load(model_path)
        # model.summary()
        logging.info("Model W2V loaded successfully.")
        return model
    else:
        logging.info("Model W2V file does not exist.")
        return None

def loadModel() -> keras.Model:
    model_path = os.path.join('./', 'text_classification_model.h5')
    if os.path.isfile(model_path):
        # Load the model if it exists
        # model = keras.model.load
        model = tf.keras.models.load_model(model_path)
        model.summary()
        logging.info("Model loaded successfully.")
        return model
    else:
        logging.info("Model file does not exist.")
        return None

if __name__ == '__main__':
    model = loadModel()
    main_model = model
    word2vec_model = loadWord2VecModel()
    w2v_model = word2vec_model
    try:
        label_encoder = joblib.load("label_encoder.pkl")
        logging.info("Model label_encoder loaded successfully.")
    except:
        label_encoder = None
    
    logging.info("Flask running on port 8081")
    app.run(debug=None, host='localhost', port=8081)

    



