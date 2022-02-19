<?php

namespace DesignPatterns\Structural\Composite\Tests;

use DesignPatterns\Structural\Composite\Fieldset;
use DesignPatterns\Structural\Composite\Form;
use DesignPatterns\Structural\Composite\Input;
use Orchestra\Testbench\TestCase;

class CompositeTest extends TestCase
{
    /** @test */
    public function is_form_object_attribute_set_properly()
    {
        $form = new Form(
            $name = 'user info',
            $title = 'Your information',
            $url = '/api/send-info');

        $this->assertEquals($name, $form->getName());
        $this->assertMatchesRegularExpression('/<h3>' . $title . '<\/h3>/', $form->render());
    }

    /** @test */
    public function can_a_form_object_add_input()
    {
        $form = new Form(
            'user info',
            'Your information',
            '/api/send-info');

        $form->add(new Input($name = 'name', 'name', $type = 'text'));

        $this->assertMatchesRegularExpression(
            '/<input name="' . $name . '" type="' . $type . '" value="">/',
            $form->render()
        );

    }

    /** @test */
    public function can_a_form_object_add_fieldset()
    {
        $form = new Form('product', "Add product", "/product/add");
        $form->add(new Input('name', "Name", 'text'));
        $form->add(new Input('description', "Description", 'text'));

        $picture = new Fieldset('photo', $title = "Product photo");
        $picture->add(new Input('caption', "Caption", 'text'));
        $picture->add(new Input('image', "Image", 'file'));
        $form->add($picture);

        $this->assertMatchesRegularExpression(
            '/<fieldset><legend>' . $title . '<\/legend>/',
            $form->render()
        );
    }
}