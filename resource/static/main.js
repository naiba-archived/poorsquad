$('.message .close').on('click', function () {
    $(this).closest('.message').transition('fade');
});

function addCompany() {
    $('.tiny.company.modal').modal({
        closable: true,
        onApprove: function () {
            let success = false
            if ($('.tiny.company.modal .positive.button').hasClass('loading')) {
                console.log('加载中，不要重复点击')
                return success
            }
            $('.tiny.company.modal .positive.button').toggleClass('loading')
            const data = $('#companyForm').serializeArray().reduce(function (obj, item) {
                obj[item.name] = item.value;
                return obj;
            }, {});
            $.post('/api/company', JSON.stringify(data)).done(function () {
                alert("second success");
            }).fail(function () {
                alert("error");
            }).always(function () {
                $('.tiny.company.modal .positive.button').toggleClass('loading')
            });
            return success
        }
    }).modal('show')
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
