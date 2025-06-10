const chatBox = document.getElementById('chat-box');
const chatForm = document.getElementById('chat-form');
const userInput = document.getElementById('user-input');

chatForm.addEventListener('submit', function(event) {
    event.preventDefault();

    const userMessageText = userInput.value.trim();

    if (userMessageText === '') {
        return;
    }

    addMessage(userMessageText, 'user');

    userInput.value = '';

    sendMessageToBackend(userMessageText);
});

function addMessage(text, sender) {
    const messageElement = document.createElement('div');
    messageElement.classList.add('message', `${sender}-message`);

    const pElement = document.createElement('p');
    pElement.textContent = text;
    messageElement.appendChild(pElement);

    chatBox.appendChild(messageElement);

    chatBox.scrollTop = chatBox.scrollHeight;
}

function sendMessageToBackend(message) {
    addMessage('...', 'bot');
    
    const apiUrl = 'https://localhost:5001/ask'; 

    fetch(apiUrl, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            message: message
        })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json();
    })
    .then(data => {
        const typingIndicator = chatBox.querySelector('.bot-message:last-child');
        typingIndicator.remove();

        const botResponse = data.reply; 
        addMessage(botResponse, 'bot');
    })
    .catch(error => {
        const typingIndicator = chatBox.querySelector('.bot-message:last-child');
        typingIndicator.remove();

        console.error('Lỗi khi gọi API:', error);
        addMessage('Xin lỗi, đã có lỗi xảy ra. Vui lòng thử lại sau.', 'bot');
    });
}