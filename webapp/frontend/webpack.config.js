const path = require('path');

module.exports = {
  "mode": "none",
  "entry": "./frontend/src/index.js",
  "output": {
    "path": path.resolve(__dirname, 'dist'),
    "filename": "js/main.js"
  },
  devServer: {
    contentBase: path.join(__dirname, "frontend", 'dist')
  },
  "module": {
    "rules": [
      {
        "test": /\.css$/,
        "use": [
          "style-loader",
          "css-loader"
        ]
      },
      {
        "test": /\.js$/,
        "exclude": /node_modules/,
        "use": {
          "loader": "babel-loader",
          "options": {
            "presets": [
              "@babel/preset-env",
            ]
          }
        }
      },
    ]
  }
}

