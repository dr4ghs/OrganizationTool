import './styles/main.css';
import JsonEnc from './json-enc';

import Alpine  from 'alpinejs';
import persist from '@alpinejs/persist';


window.Alpine = Alpine;

document.addEventListener("DOMContentLoaded", () => {
  Alpine.plugin(persist);
  Alpine.start();

  JsonEnc();

  class Activities extends Array {
    init = () => {
      let stored = localStorage.getItem("activities");
      if (!stored)
        return;

      return JSON.parse(stored).forEach(x => {
        let a = this.add(x.id, x.name, x.user, x.points, x.goal, x.type);
        a.save();
      });
    }

    #save = (a) => {
      let id = this.findIndex(x => x.id === a.id);
      if (id != -1)
        this[id] = a;
      else
        this.push(a);
    }

    #discard = (a) => {
      let id = this.findIndex(x => x.id === a.id);
      if (id !== -1) {
        this.splice(id, 1);
        document.getElementById('a_' + a.id)?.remove();
      }
    }

    #update = (func, a) => {
      func(a);

      localStorage.setItem("activities", JSON.stringify(this));
    }

    get = (id) => this.filter(x => x.id === id)[0]
    add = (id, name, user, points, goal, type) => {
      let a = {
        id: id,
        name: name,
        user: user,
        points: points,
        goal: goal,
        type: type,

        save: () => this.#update(this.#save, a),
        discard: () => this.#update(this.#discard, a),
      };

      return a;
    }
    new = () => {
      let a = this.add(
        "newActivity-" + this.length,
        "New Activity",
        JSON.parse(localStorage.getItem("user")).id,
        0,
        0,
        "daily"
      );
      a.save();

      return a;
    }
  }


  const OrgTool = {
    _activities: null,

    get activities() {
      if (!this._activities) {
        console.log("NEW")

        this._activities = new Activities();
        this._activities.init();
      }

      return this._activities;
    },

    login: (identity, password) => fetch("/api/auth/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json; charset=UTF-8"
        },
        body: JSON.stringify({
          identity: identity,
          password: password,
        }),
      })
      .then((res) => {
        if (res.ok) {
          return res.json();
        }

        return Promise.reject(res);
      })
      .then((json) => {
        localStorage.setItem("user", JSON.stringify(json));
        Alpine.store("user", json)
      }),
    logout: () => fetch("/api/auth/logout", {}).then(() => localStorage.removeItem("user")),
    user: () => {
      if (localStorage.getItem("user")) {
        Alpine.store("user", JSON.parse(localStorage.getItem("user")))
      }

      return Alpine.store("user")
    },
  }

  window.OrgTool = OrgTool
});

