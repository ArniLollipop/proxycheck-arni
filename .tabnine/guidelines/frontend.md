# 🧩 Frontend Coding Guidelines — Vue 2 + BootstrapVue + Pug

## 📁 Project Overview
- All frontend code lives in `/client`.
- Framework: **Vue 2**
- UI Library: **BootstrapVue**
- Template Engine: **Pug**
- **No Vuex** — use local component state or simple helper modules.
- **No TypeScript** — only plain JavaScript (ES6+).
- **Minimal custom CSS** — prefer Bootstrap utility classes.

---

## 🧱 Project Structure
/client
├── /components # Reusable UI components (tables, forms, modals, etc.)
├── /views # Page-level components
├── /assets # Images, icons, SCSS
├── /plugins # BootstrapVue, Axios, etc.
├── /router # Vue Router (optional)
├── App.vue
└── main.js

---

## 🧠 Core Coding Rules

### 1. Components
- Located in `/client/components`.
- File names use **PascalCase.vue**.
- Example structure:

```vue
<template lang="pug">
div.my-component
  b-card
    b-table(:items="items" :fields="fields")
</template>

<script>
export default {
  name: 'MyComponent',
  props: {},
  data() {
    return {}
  },
  methods: {}
}
</script>
2. Views (Pages)
Located in /client/views.
Use Pug templates.
Import and combine components.
Example:
<template lang="pug">
b-container
  h3.text-center Users List
  UserTable
</template>

<script>
import UserTable from '@/components/UserTable.vue'

export default {
  name: 'UsersView',
  components: { UserTable }
}
</script>
3. Complex Tables (Main Feature)
Use <b-table> from BootstrapVue.
Support sorting, filtering, and pagination.
Do not use third-party grid libraries.
Example:
<template lang="pug">
b-container(fluid)
  b-row
    b-col(lg="6" class="my-1")
      b-form-group(label="Filter" label-cols-sm="3" label-size="sm")
        b-input-group(size="sm")
          b-form-input(v-model="filter" placeholder="Type to Search")
          b-input-group-append
            b-button(:disabled="!filter" @click="filter = ''") Clear
  b-table(
    :items="items"
    :fields="fields"
    :filter="filter"
    :sort-by.sync="sortBy"
    :sort-desc.sync="sortDesc"
    show-empty
    small
  )
    template(#cell(actions)="row")
      b-button(size="sm" @click="info(row.item)") Info
</template>

<script>
export default {
  data() {
    return {
      filter: '',
      sortBy: 'age',
      sortDesc: false,
      items: [
        { name: 'Alice', age: 30, isActive: true },
        { name: 'Bob', age: 25, isActive: false }
      ],
      fields: [
        { key: 'name', label: 'Name', sortable: true },
        { key: 'age', label: 'Age', sortable: true },
        { key: 'isActive', label: 'Active', formatter: v => (v ? 'Yes' : 'No') },
        { key: 'actions', label: 'Actions' }
      ]
    }
  },
  methods: {
    info(item) {
      alert(JSON.stringify(item, null, 2))
    }
  }
}
</script>
4. Forms and Filters
Use b-form, b-form-group, and b-input-group.
Keep logic simple — no external filtering libraries.
Example:
b-form-input(v-model="search" type="search" placeholder="Search by name")
b-table(:items="filteredItems")
5. Modals
Use only b-modal from BootstrapVue.
Open modals via:
this.$root.$emit('bv::show::modal', modalId)
6. Template Rules
Avoid logic-heavy templates.
All calculations belong in computed or methods.
Templates should only contain rendering and control structures (v-if, v-for, etc.).
7. General Principles
✅ Simplicity over abstraction
✅ Prefer BootstrapVue utilities over custom CSS
✅ Explicit props and data
✅ Clean, indented Pug structure
✅ No Vuex / Pinia / Composition API
8. Example Page Structure
/client/views/UsersView.vue
/client/components/UserTable.vue
/client/components/UserInfoModal.vue
9. Tabnine / AI Assistant Style Rules
AI-generated code should:
Use lang="pug" templates.
Follow Vue 2 syntax (not Vue 3).
Use BootstrapVue components (b- prefix).
Avoid TypeScript, Vuex, and Composition API.
Keep data() sections minimal and explicit.
Avoid unnecessary custom CSS.
Focus on tables, filters, and simple UI logic.
✅ TL;DR Summary
Category	Use	Avoid
Framework	Vue 2	Vue 3
UI	BootstrapVue	Vuetify / ElementUI
Templates	Pug	Raw HTML
State Management	Local data	Vuex / Pinia
CSS	Bootstrap utilities	Custom CSS
Language	JavaScript	TypeScript
Features	Tables, filters, modals	Complex animations, external grids

---