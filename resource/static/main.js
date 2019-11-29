$('.ui.checkbox').checkbox();
$('.ui.dropdown').dropdown();

function showConfirm(title, content, callFn, extData) {
    const modal = $('.mini.confirm.modal')
    modal.children('.header').text(title)
    modal.children('.content').text(content)
    modal.modal({
        closable: true,
        onApprove: function () {
            callFn(extData)
            return true
        }
    }).modal('show')
}

function showFormModal(modelSelector, formID, URL, getData) {
    $(modelSelector).modal({
        closable: true,
        onApprove: function () {
            let success = false
            const btn = $(modelSelector + ' .positive.button')
            const form = $(modelSelector + ' form')
            if (btn.hasClass('loading')) {
                return success
            }
            form.children('.message').remove()
            btn.toggleClass('loading')
            const data = getData ? getData() : $(formID).serializeArray().reduce(function (obj, item) {
                obj[item.name] = (item.name.endsWith('_id') || item.name === 'id' || item.name === 'permission') ? parseInt(item.value) : item.value;
                return obj;
            }, {});
            $.post(URL, JSON.stringify(data)).done(function (resp) {
                if (resp.code == 200) {
                    if (resp.message) {
                        alert(resp.message)
                    }
                    window.location.reload()
                } else {
                    form.append(`<div class="ui negative message"><div class="header">操作失败</div><p>` + resp.message + `</p></div>`)
                }
            }).fail(function (err) {
                form.append(`<div class="ui negative message"><div class="header">网络错误</div><p>` + err.responseText + `</p></div>`)
            }).always(function () {
                btn.toggleClass('loading')
            });
            return success
        }
    }).modal('show')
}

function addTeam() {
    showFormModal('.tiny.team.modal', '#teamForm', '/api/team');
}

function bindRepository(id, repos) {
    $('#bindRepositoryForm input[name=id]').val(id)
    $('#bindRepositoryForm .checkbox').checkbox('uncheck')
    if (repos) {
        for (let i = 0; i < repos.length; i++) {
            $('#bindRepositoryForm .id-' + repos[i]).checkbox('check')
        }
    }
    getData = function () {
        return $('#bindRepositoryForm').serializeArray().reduce(function (obj, item) {
            if (!obj['repositories']) {
                obj['repositories'] = []
            }
            if (item.value === 'on') {
                obj['repositories'].push(parseInt(item.name))
            } else {
                obj[item.name] = (item.name.endsWith('_id') || item.name === 'id') ? parseInt(item.value) : item.value;
            }
            return obj;
        }, {})
    }
    showFormModal('.tiny.bind-repository.modal', '#bindRepositoryForm', '/api/team/repositories', getData)
}

function addAccount() {
    showFormModal('.tiny.account.modal', '#accountForm', '/api/account');
}

function addCompany() {
    showFormModal('.tiny.company.modal', '#companyForm', '/api/company');
}

function addEmployee(type, id) {
    $('#employeeForm .dropdown .item:nth-child(3)').css('display', 'block')
    $('#employeeForm .dropdown').parent().css('display', 'block')
    $('#employeeForm input[name=id]').val(id)
    $('#employeeForm input[name=type]').val(type)
    $('#employeeForm .dropdown').dropdown('set selected', 1)
    if (type === 'repositoryOutsideCollaborator') {
        $('#employeeForm .dropdown').parent().css('display', 'none')
    } else if (type === 'team') {
        $('#employeeForm .dropdown .item:nth-child(3)').css('display', 'none')
    }
    showFormModal('.tiny.employee.modal', '#employeeForm', '/api/employee');
}

function removeEmployee(data) {
    $.ajax({
        url: '/api/employee/' + data.type + '/' + data.id + '/' + data.userID,
        type: 'DELETE',
    }).done(resp => {
        if (resp.code == 200) {
            if (resp.message) {
                alert(resp.message)
            } else {
                alert('移出成功')
            }
            window.location.reload()
        } else {
            alert('移出失败 ' + resp.code + '：' + resp.message)
        }
    }).fail(err => {
        lert('网络错误：' + err.responseText)
    });
}

function logout(id) {
    $.post('/api/logout', JSON.stringify({ id: id })).done(function (resp) {
        if (resp.code == 200) {
            alert('注销成功')
            window.location.href = '/login'
        } else {
            alert('注销失败 ' + resp.code + '：' + resp.message)
        }
    }).fail(function (err) {
        alert('网络错误：' + err.responseText)
    })
}
