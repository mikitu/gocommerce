import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';
import { connect } from 'react-redux';
import { simpleAction } from './actions/simpleAction'
import Button from '@material-ui/core/Button';
const mapDispatchToProps = dispatch => ({
    simpleAction: () => dispatch(simpleAction())
})

const mapStateToProps = state => ({
    ...state
})

class App extends Component {
    simpleAction = (event) => {
        this.props.simpleAction();
    }

    render() {
    return (
      <div className="App">
        <header className="App-header">
          <img src={logo} className="App-logo" alt="logo" />
          <p>
            Edit <code>src/App.js</code> and save to reload.
          </p>
          <a
            className="App-link"
            href="https://reactjs.org"
            target="_blank"
            rel="noopener noreferrer"
          >
            Learn React
          </a>
        </header>
          <Button variant="contained" color="primary" onClick={this.simpleAction}>
              Hello World
          </Button>
          <pre>
             {
                 JSON.stringify(this.props)
             }
            </pre>
        </div>
    );
  }
}


export default connect(mapStateToProps, mapDispatchToProps)(App);
