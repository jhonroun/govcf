package main

import (
	"encoding/base64"
	"html/template"
	"os"
)

type ContactReport struct {
	Total         int
	WithPhones    int
	WithoutPhones int
	Contacts      []Contact
}

func generateHTMLReport(contacts []Contact, filename string) error {
	tpl := `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<title>Контакты</title>
<style>
/* Bootstrap 5.3.0 core styles (минимально необходимые) */
body {
  font-family: -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;
  font-size: 1rem;
  line-height: 1.5;
  margin: 0;
  padding: 1rem;
  max-width: 100vw;
  overflow-x: auto;
  box-sizing: border-box;
}
.table { width: 100%; margin-bottom: 1rem; color: #212529; border-collapse: collapse; }
.table th, .table td { padding: .75rem; vertical-align: top; border-top: 1px solid #dee2e6; }
.table-bordered { border: 1px solid #dee2e6; }
.table-bordered th, .table-bordered td { border: 1px solid #dee2e6; }
.table-striped tbody tr:nth-of-type(odd) { background-color: rgba(0,0,0,.05); }
.table-hover tbody tr:hover { background-color: rgba(0,0,0,.075); }
.table-dark th { background-color: #343a40; color: white; }
.form-control { display: block; width: 100%; padding: .375rem .75rem; font-size: 1rem; line-height: 1.5; color: #495057; background-color: #fff; border: 1px solid #ced4da; border-radius: .25rem; }
img.thumb { cursor: pointer; max-height: 80px; }
.modal { display: none; position: fixed; z-index: 1050; top: 0; left: 0; width: 100%; height: 100%; overflow: hidden; background-color: rgba(0,0,0,.5); }
.modal-dialog { position: relative; margin: 1.75rem auto; max-width: 500px; }
.modal-content { position: relative; display: flex; flex-direction: column; background-color: #fff; border: 1px solid rgba(0,0,0,.2); border-radius: .3rem; padding: 1rem; }
.modal-header { display: flex; justify-content: space-between; align-items: center; }
.modal-body { position: relative; flex: 1 1 auto; padding: 1rem; text-align: center; }
.modal.show { display: flex; justify-content: center; align-items: center; }
button#printBtn { margin-bottom: 1rem; padding: 0.5rem 1rem; background: #007bff; color: white; border: none; border-radius: 0.25rem; cursor: pointer; }
button#printBtn:hover { background: #0056b3; }
@media print {
  button#printBtn, #searchInput {
    display: none;
  }
  body {
    width: 100%%;
    height: 100%%;
    margin: 0;
  }
  @page {
    size: landscape;
  }
}
.d-flex { display: flex; }
.justify-content-between { justify-content: space-between; }
.align-items-center { align-items: center; }
.flex-wrap { flex-wrap: wrap; }
.mb-3 { margin-bottom: 1rem; }
.theme-toggle {
  margin-bottom: 1rem;
  padding: 0.5rem 1rem;
  background: #6c757d;
  color: white;
  border: none;
  border-radius: 0.25rem;
  cursor: pointer;
}
.theme-toggle:hover {
  background: #5a6268;
}
.dark-mode {
  background-color: #121212;
  color: #e0e0e0;
}
.dark-mode .table {
  color: #e0e0e0;
}
.dark-mode .table th,
.dark-mode .table td {
  border-color: #333;
}
.dark-mode .table-dark th {
  background-color: #1f1f1f;
}
.dark-mode .modal-content {
  background-color: #2c2c2c;
  color: white;
}
h1.report-title {
  text-align: center;
  font-size: 2.2rem;
  font-weight: bold;
  margin-bottom: 2rem;
  color: #444;
  text-shadow: 1px 1px 2px rgba(0,0,0,0.1);
}
/* Прелоадер */
#preloader {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  opacity: 1;
  transition: opacity 0.5s ease;
  pointer-events: none;
}
#preloader.hidden {
  opacity: 0;
}
#preloader svg {
  width: 64px;
  height: 64px;
  animation: spin 2s linear infinite;
}
@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
@media print {
  .theme-toggle {
    display: none;
  }
}
</style>
<script>
document.addEventListener('DOMContentLoaded', function () {
  const modal = document.getElementById('photoModal');
  const modalImage = document.getElementById('modalImage');

  document.querySelectorAll('img.thumb').forEach(img => {
    img.addEventListener('click', () => {
      modalImage.src = img.getAttribute('data-bs-image');
      modal.classList.add('show');
    });
  });

  document.querySelector('#photoModal .btn-close').addEventListener('click', () => {
    modal.classList.remove('show');
  });

  document.getElementById('searchInput').addEventListener('keyup', function () {
    const filter = this.value.toLowerCase();
    document.querySelectorAll('#contactsTable tr').forEach(row => {
      const text = row.textContent.toLowerCase();
      row.style.display = text.includes(filter) ? '' : 'none';
    });
  });

  // Сортировка по столбцам
  document.querySelectorAll('th').forEach((th, index) => {
    th.style.cursor = 'pointer';
    th.addEventListener('click', () => {
      const rows = Array.from(document.querySelectorAll('#contactsTable tr'));
      const asc = th.classList.toggle('asc');
      th.classList.toggle('desc', !asc);
      rows.sort((a, b) => {
        const aText = a.cells[index]?.innerText || "";
        const bText = b.cells[index]?.innerText || "";
        return asc ? aText.localeCompare(bText) : bText.localeCompare(aText);
      });
      rows.forEach(r => document.getElementById('contactsTable').appendChild(r));
    });
  });

  // Фильтр по наличию телефона
  const telFilter = document.getElementById('telFilter');
  telFilter.addEventListener('change', () => {
    document.querySelectorAll('#contactsTable tr').forEach(row => {
      const hasPhone = row.cells[3]?.innerText.trim() !== "";
      row.style.display = telFilter.checked ? (hasPhone ? '' : 'none') : '';
    });
  });
});
// Тема
window.toggleTheme = function() {
  document.body.classList.toggle('dark-mode');
}
</script>
<style>
  img.thumb { cursor: pointer; max-height: 80px; }
</style>
</head>
<body class="p-4">
<h1 class="report-title">LemTech (c) 2025 Парсер VCF</h1>
<h1 class="mb-4">Список контактов</h1>
<div class="d-flex justify-content-between align-items-center flex-wrap mb-3" style="gap: 1rem;">
 <div class="mb-4">
	<strong>Всего контактов:</strong> {{.Total}}<br>
	<strong>С телефонами:</strong> {{.WithPhones}}<br>
	<strong>Без телефонов:</strong> {{.WithoutPhones}}
	</div>
  <input class="form-control" id="searchInput" type="text" placeholder="Поиск по контактам..." style="max-width: 400px; flex: 1;">
  <label style="display:flex;align-items:center;gap:0.5rem;font-size:0.9rem;">
    <input type="checkbox" id="telFilter"> Только с телефонами
  </label>
  <button class="theme-toggle" onclick="toggleTheme()">🌓 Переключить тему</button>
  <button id="printBtn" onclick="window.print()">Сохранить как PDF</button>
</div>
<table class="table table-bordered table-striped table-hover">
<thead class="table-dark">
<tr>
<th>№</th><th>Полное имя</th><th>Имя (структура)</th><th>Телефоны</th><th>Email</th><th>Организация</th><th>Должность</th><th>Адрес</th><th>URL</th><th>Заметки</th><th>Фото</th>
</tr>
</thead>
<tbody id="contactsTable">
{{range .Contacts}}
<tr>
<td>{{.Index}}</td>
<td>{{.FN}}</td>
<td>{{range .N}}{{.}} {{end}}</td>
<td>{{range .TEL}}{{.}}<br>{{end}}</td>
<td>{{range .EMAIL}}{{.}}<br>{{end}}</td>
<td>{{range .ORG}}{{.}}<br>{{end}}</td>
<td>{{.TITLE}}</td>
<td>{{range .ADR}}{{.}}<br>{{end}}</td>
<td>{{range .URL}}{{.}}<br>{{end}}</td>
<td>{{range .NOTE}}{{.}}<br>{{end}}</td>
<td>
{{if .PHOTO}}
<img src="data:image/jpeg;base64,{{base64 .PHOTO}}" class="thumb" data-bs-toggle="modal" data-bs-target="#photoModal" data-bs-image="data:image/jpeg;base64,{{base64 .PHOTO}}">
{{else}}—{{end}}
</td>
</tr>
{{end}}
</tbody>
</table>

<!-- Modal -->
<div id="photoModal" class="modal" role="dialog">
  <div class="modal-dialog">
    <div class="modal-content">
      <div class="modal-header">
        <h5 class="modal-title">Фото</h5>
        <button type="button" class="btn-close" aria-label="Закрыть" onclick="document.getElementById('photoModal').classList.remove('show')">×</button>
      </div>
      <div class="modal-body text-center">
        <img id="modalImage" class="img-fluid" alt="Фото контакта">
      </div>
    </div>
  </div>
</div>
</div>

<script>
document.querySelectorAll('img.thumb').forEach(img => {
  img.addEventListener('click', () => {
    document.getElementById('modalImage').src = img.getAttribute('data-bs-image');
  });
});

document.getElementById('searchInput').addEventListener('keyup', function () {
  const filter = this.value.toLowerCase();
  document.querySelectorAll('#contactsTable tr').forEach(row => {
    const text = row.textContent.toLowerCase();
    row.style.display = text.includes(filter) ? '' : 'none';
  });
});
</script>
</body>
</html>`

	t := template.New("report").Funcs(template.FuncMap{
		"base64": func(data []byte) string {
			return base64.StdEncoding.EncodeToString(data)
		},
	})
	t, err := t.Parse(tpl)
	if err != nil {
		return err
	}

	var withPhones int
	for _, c := range contacts {
		if len(c.TEL) > 0 {
			withPhones++
		}
	}

	reportData := ContactReport{
		Total:         len(contacts),
		WithPhones:    withPhones,
		WithoutPhones: len(contacts) - withPhones,
		Contacts:      contacts,
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, reportData)
}
